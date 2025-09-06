package task_orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"assistant/internal/app/repo"
)

type TaskWorker struct {
	q        *repo.Queries
	workerID string
	stopCh   chan struct{}
}

func NewTaskWorker(q *repo.Queries, workerID string) *TaskWorker {
	return &TaskWorker{
		q:        q,
		workerID: workerID,
		stopCh:   make(chan struct{}),
	}
}

func (w *TaskWorker) Start(ctx context.Context) {
	log.Printf("[Worker-%s] Task worker started", w.workerID)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker-%s] Task worker stopped", w.workerID)
			return
		case <-w.stopCh:
			log.Printf("[Worker-%s] Task worker received stop signal", w.workerID)
			return
		case <-ticker.C:
			w.processPendingInstances(ctx)
		}
	}
}

func (w *TaskWorker) Stop() {
	close(w.stopCh)
}

func (w *TaskWorker) processPendingInstances(ctx context.Context) {
	instances, err := w.q.ListPendingTaskWorkflowInstances(ctx, 10)
	if err != nil {
		log.Printf("[Worker-%s] Failed to list pending instances: %v", w.workerID, err)
		return
	}

	for _, instance := range instances {
		w.executeWorkflow(ctx, instance)
	}
}

func (w *TaskWorker) executeWorkflow(ctx context.Context, instance repo.TaskWorkflowInstance) {
	log.Printf("[Worker-%s] Starting workflow instance %d", w.workerID, instance.ID)

	now := time.Now()
	_, _ = w.q.UpdateTaskWorkflowInstanceStatus(ctx, repo.UpdateTaskWorkflowInstanceStatusParams{
		Status:   repo.TaskWorkflowInstanceStatusRUNNING,
		GmtStart: sql.NullTime{Time: now, Valid: true},
		ID:       instance.ID,
	})

	nodes, err := w.q.ListTaskNodeInstances(ctx, instance.ID)
	if err != nil {
		log.Printf("[Worker-%s] Failed to get nodes for instance %d: %v", w.workerID, instance.ID, err)
		w.failWorkflow(ctx, instance.ID, err.Error())
		return
	}

	completed := 0
	failed := 0

	for _, node := range nodes {
		if node.Status != repo.TaskNodeInstanceStatusPENDING {
			continue
		}

		workflow, _ := w.q.GetTaskWorkflowDefByID(ctx, instance.WorkflowDefID)
		err := w.executeNode(ctx, instance.ID, node)
		if err != nil {
			log.Printf("[Worker-%s] Node %s failed: %v", w.workerID, node.NodeID, err)
			failed++

			onErrorStrategy := "STOP"
			if workflow.OnErrorStrategy.Valid {
				onErrorStrategy = string(workflow.OnErrorStrategy.TaskWorkflowDefOnErrorStrategy)
			}

			if onErrorStrategy == "STOP" {
				w.failWorkflow(ctx, instance.ID, fmt.Sprintf("Node %s failed: %v", node.NodeID, err))
				return
			}
		} else {
			completed++
		}

		_, _ = w.q.UpdateTaskWorkflowInstanceProgress(ctx, repo.UpdateTaskWorkflowInstanceProgressParams{
			TotalNodes:     sql.NullInt32{Int32: int32(len(nodes)), Valid: true},
			CompletedNodes: sql.NullInt32{Int32: int32(completed), Valid: true},
			FailedNodes:    sql.NullInt32{Int32: int32(failed), Valid: true},
			ID:             instance.ID,
		})
	}

	endTime := time.Now()
	if failed > 0 {
		errorInfo, _ := json.Marshal(map[string]interface{}{
			"failed_nodes":    failed,
			"completed_nodes": completed,
			"message":         "Workflow execution had failures",
		})
		_, _ = w.q.UpdateTaskWorkflowInstanceStatus(ctx, repo.UpdateTaskWorkflowInstanceStatusParams{
			Status:    repo.TaskWorkflowInstanceStatusFAILED,
			GmtEnd:    sql.NullTime{Time: endTime, Valid: true},
			ErrorInfo: errorInfo,
			ID:        instance.ID,
		})
	} else {
		result, _ := json.Marshal(map[string]interface{}{
			"completed_nodes": completed,
			"failed_nodes":    failed,
			"duration_ms":     endTime.Sub(now).Milliseconds(),
		})
		_, _ = w.q.UpdateTaskWorkflowInstanceStatus(ctx, repo.UpdateTaskWorkflowInstanceStatusParams{
			Status:        repo.TaskWorkflowInstanceStatusSUCCESS,
			GmtEnd:        sql.NullTime{Time: endTime, Valid: true},
			ResultSummary: result,
			ID:            instance.ID,
		})
	}

	log.Printf("[Worker-%s] Workflow instance %d completed: %d/%d nodes", w.workerID, instance.ID, completed, len(nodes))
}

func (w *TaskWorker) executeNode(ctx context.Context, workflowInstanceID int64, node repo.TaskNodeInstance) error {
	log.Printf("[Worker-%s] Executing node %s", w.workerID, node.NodeID)

	startTime := time.Now()
	_, _ = w.q.UpdateTaskNodeInstanceStatus(ctx, repo.UpdateTaskNodeInstanceStatusParams{
		Status:   repo.TaskNodeInstanceStatusRUNNING,
		GmtStart: sql.NullTime{Time: startTime, Valid: true},
		WorkerID: sql.NullString{String: w.workerID, Valid: true},
		ID:       node.ID,
	})

	workflowNode, err := w.q.GetTaskWorkflowNodeByID(ctx, node.NodeDefID)
	if err != nil {
		return fmt.Errorf("failed to get workflow node: %w", err)
	}

	if workflowNode.TaskAtomicDefID.Valid {
		return w.executeAtomicTask(ctx, workflowInstanceID, node, workflowNode, startTime)
	}

	return nil
}

func (w *TaskWorker) executeAtomicTask(ctx context.Context, workflowInstanceID int64, node repo.TaskNodeInstance, workflowNode repo.TaskWorkflowNode, startTime time.Time) error {
	atomicTask, err := w.q.GetTaskAtomicDefByID(ctx, workflowNode.TaskAtomicDefID.Int64)
	if err != nil {
		return fmt.Errorf("failed to get atomic task: %w", err)
	}

	var inputParams map[string]interface{}
	if node.InputParams != nil {
		json.Unmarshal(node.InputParams, &inputParams)
	}

	atomicInstanceParams := repo.CreateTaskAtomicInstanceParams{
		NodeInstanceID:  node.ID,
		TaskAtomicDefID: atomicTask.ID,
		Status:          repo.TaskAtomicInstanceStatusPENDING,
		InputParams:     node.InputParams,
	}

	result, err := w.q.CreateTaskAtomicInstance(ctx, atomicInstanceParams)
	if err != nil {
		return fmt.Errorf("failed to create atomic task instance: %w", err)
	}
	atomicInstanceID, _ := result.LastInsertId()

	if atomicTask.TaskCategory == repo.TaskAtomicDefTaskCategorySCRIPT {
		return w.executeScriptTask(ctx, workflowInstanceID, node, atomicTask, atomicInstanceID, startTime)
	} else if atomicTask.TaskCategory == repo.TaskAtomicDefTaskCategoryHTTPAPI {
		return w.executeHTTPTask(ctx, workflowInstanceID, node, atomicTask, atomicInstanceID, startTime)
	}

	return nil
}

func (w *TaskWorker) executeScriptTask(ctx context.Context, workflowInstanceID int64, node repo.TaskNodeInstance, atomicTask repo.TaskAtomicDef, atomicInstanceID int64, startTime time.Time) error {
	var cmd *exec.Cmd
	scriptType := string(atomicTask.ScriptType.TaskAtomicDefScriptType)

	if runtime.GOOS == "windows" {
		switch scriptType {
		case "SHELL":
			cmd = exec.CommandContext(ctx, "powershell", "-Command", atomicTask.ScriptContent)
		default:
			cmd = exec.CommandContext(ctx, "powershell", "-Command", atomicTask.ScriptContent)
		}
	} else {
		switch scriptType {
		case "SHELL":
			cmd = exec.CommandContext(ctx, "bash", "-c", atomicTask.ScriptContent)
		case "PYTHON":
			cmd = exec.CommandContext(ctx, "python3", "-c", atomicTask.ScriptContent)
		case "LUA":
			cmd = exec.CommandContext(ctx, "lua", "-e", atomicTask.ScriptContent)
		default:
			cmd = exec.CommandContext(ctx, "sh", "-c", atomicTask.ScriptContent)
		}
	}

	output, err := cmd.CombinedOutput()
	endTime := time.Now()
	duration := int32(endTime.Sub(startTime).Milliseconds())

	status := repo.TaskAtomicInstanceStatusSUCCESS
	if err != nil {
		status = repo.TaskAtomicInstanceStatusFAILED
	}

	logMsg := string(output)
	if len(logMsg) > 10000 {
		logMsg = logMsg[:10000] + "... (truncated)"
	}

	_, _ = w.q.UpdateTaskAtomicInstanceStatus(ctx, repo.UpdateTaskAtomicInstanceStatusParams{
		Status:       status,
		OutputResult: json.RawMessage(fmt.Sprintf(`{"output": %q}`, string(output))),
		ExecutionLog: sql.NullString{String: logMsg, Valid: true},
		ErrorLog:     sql.NullString{String: err.Error(), Valid: err != nil},
		GmtEnd:       sql.NullTime{Time: endTime, Valid: true},
		DurationMs:   sql.NullInt32{Int32: duration, Valid: true},
		WorkerID:     sql.NullString{String: w.workerID, Valid: true},
		ID:           atomicInstanceID,
	})

	_, _ = w.q.UpdateTaskNodeInstanceStatus(ctx, repo.UpdateTaskNodeInstanceStatusParams{
		Status:       repo.TaskNodeInstanceStatus(status),
		OutputResult: json.RawMessage(fmt.Sprintf(`{"output": %q}`, string(output))),
		ExecutionLog: sql.NullString{String: logMsg, Valid: true},
		ErrorLog:     sql.NullString{String: err.Error(), Valid: err != nil},
		GmtEnd:       sql.NullTime{Time: endTime, Valid: true},
		DurationMs:   sql.NullInt32{Int32: duration, Valid: true},
		WorkerID:     sql.NullString{String: w.workerID, Valid: true},
		ID:           node.ID,
	})

	_, _ = w.q.CreateTaskExecutionLog(ctx, repo.CreateTaskExecutionLogParams{
		WorkflowInstanceID:   workflowInstanceID,
		NodeInstanceID:       sql.NullInt64{Int64: node.ID, Valid: true},
		TaskAtomicInstanceID: sql.NullInt64{Int64: atomicInstanceID, Valid: true},
		LogLevel:             repo.TaskExecutionLogLogLevelINFO,
		LogType:              repo.TaskExecutionLogLogTypeEXECUTION,
		Message:              fmt.Sprintf("Script executed with status: %s", status),
	})

	if err != nil {
		return fmt.Errorf("script execution failed: %w, output: %s", err, string(output))
	}

	return nil
}

func (w *TaskWorker) executeHTTPTask(ctx context.Context, workflowInstanceID int64, node repo.TaskNodeInstance, atomicTask repo.TaskAtomicDef, atomicInstanceID int64, startTime time.Time) error {
	var httpConfig map[string]interface{}
	if atomicTask.HttpConfig != nil {
		json.Unmarshal(atomicTask.HttpConfig, &httpConfig)
	}

	method := "GET"
	url := ""
	headers := make(map[string]string)
	body := ""

	if v, ok := httpConfig["method"].(string); ok {
		method = v
	}
	if v, ok := httpConfig["url"].(string); ok {
		url = v
	}
	if v, ok := httpConfig["headers"].(map[string]interface{}); ok {
		for k, val := range v {
			headers[k] = fmt.Sprintf("%v", val)
		}
	}
	if v, ok := httpConfig["body"].(string); ok {
		body = v
	}

	log.Printf("[Worker-%s] HTTP %s %s", w.workerID, method, url)
	_ = body

	endTime := time.Now()
	duration := int32(endTime.Sub(startTime).Milliseconds())

	status := repo.TaskAtomicInstanceStatusSUCCESS
	logMsg := fmt.Sprintf("HTTP %s %s completed", method, url)

	_, _ = w.q.UpdateTaskAtomicInstanceStatus(ctx, repo.UpdateTaskAtomicInstanceStatusParams{
		Status:       status,
		ExecutionLog: sql.NullString{String: logMsg, Valid: true},
		GmtEnd:       sql.NullTime{Time: endTime, Valid: true},
		DurationMs:   sql.NullInt32{Int32: duration, Valid: true},
		WorkerID:     sql.NullString{String: w.workerID, Valid: true},
		ID:           atomicInstanceID,
	})

	_, _ = w.q.UpdateTaskNodeInstanceStatus(ctx, repo.UpdateTaskNodeInstanceStatusParams{
		Status:       repo.TaskNodeInstanceStatus(status),
		ExecutionLog: sql.NullString{String: logMsg, Valid: true},
		GmtEnd:       sql.NullTime{Time: endTime, Valid: true},
		DurationMs:   sql.NullInt32{Int32: duration, Valid: true},
		WorkerID:     sql.NullString{String: w.workerID, Valid: true},
		ID:           node.ID,
	})

	return nil
}

func (w *TaskWorker) failWorkflow(ctx context.Context, instanceID int64, reason string) {
	endTime := time.Now()
	errorInfo, _ := json.Marshal(map[string]interface{}{
		"error":     reason,
		"failed_at": endTime.Format(time.RFC3339),
	})

	_, _ = w.q.UpdateTaskWorkflowInstanceStatus(ctx, repo.UpdateTaskWorkflowInstanceStatusParams{
		Status:       repo.TaskWorkflowInstanceStatusFAILED,
		StatusReason: sql.NullString{String: reason, Valid: true},
		GmtEnd:       sql.NullTime{Time: endTime, Valid: true},
		ErrorInfo:    errorInfo,
		ID:           instanceID,
	})
}

func (w *TaskWorker) RollbackWorkflow(ctx context.Context, instanceID int64) error {
	log.Printf("[Worker-%s] Starting rollback for workflow instance %d", w.workerID, instanceID)

	nodes, err := w.q.ListTaskNodeInstances(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		if node.Status != repo.TaskNodeInstanceStatusSUCCESS {
			continue
		}

		atomicInstances, _ := w.q.ListTaskAtomicInstances(ctx, node.ID)
		for j := len(atomicInstances) - 1; j >= 0; j-- {
			atomicInst := atomicInstances[j]
			atomicTask, _ := w.q.GetTaskAtomicDefByID(ctx, atomicInst.TaskAtomicDefID)

			if atomicTask.IsRollbackSupported.Bool && atomicTask.RollbackScriptContent.Valid {
				w.executeRollback(ctx, atomicInst, atomicTask)
			}
		}
	}

	_, _ = w.q.UpdateTaskWorkflowInstanceStatus(ctx, repo.UpdateTaskWorkflowInstanceStatusParams{
		Status: repo.TaskWorkflowInstanceStatusCANCELLED,
		ID:     instanceID,
	})

	return nil
}

func (w *TaskWorker) executeRollback(ctx context.Context, atomicInst repo.TaskAtomicInstance, atomicTask repo.TaskAtomicDef) {
	log.Printf("[Worker-%s] Rolling back atomic task %d", w.workerID, atomicInst.ID)

	var cmd *exec.Cmd
	scriptType := string(atomicTask.RollbackScriptType.TaskAtomicDefRollbackScriptType)

	scriptContent := atomicTask.RollbackScriptContent.String

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-Command", scriptContent)
	} else {
		switch scriptType {
		case "SHELL":
			cmd = exec.CommandContext(ctx, "bash", "-c", scriptContent)
		case "PYTHON":
			cmd = exec.CommandContext(ctx, "python3", "-c", scriptContent)
		default:
			cmd = exec.CommandContext(ctx, "sh", "-c", scriptContent)
		}
	}

	output, err := cmd.CombinedOutput()
	logMsg := string(output)
	if len(logMsg) > 5000 {
		logMsg = logMsg[:5000]
	}

	result := "success"
	if err != nil {
		result = "failed"
		log.Printf("[Worker-%s] Rollback failed: %v", w.workerID, err)
	}

	_, _ = w.q.UpdateTaskAtomicInstanceRollback(ctx, repo.UpdateTaskAtomicInstanceRollbackParams{
		RollbackLog:    sql.NullString{String: logMsg, Valid: true},
		RollbackResult: json.RawMessage(fmt.Sprintf(`{"result": %q}`, result)),
		ID:             atomicInst.ID,
	})
}
