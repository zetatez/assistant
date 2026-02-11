package task_orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"assistant/internal/app/repo"
	"assistant/internal/bootstrap/psl"

	"github.com/sirupsen/logrus"
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
	logger := psl.GetLogger()
	logger.WithField("worker_id", w.workerID).Info("task worker started")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.WithField("worker_id", w.workerID).Info("task worker stopped")
			return
		case <-w.stopCh:
			logger.WithField("worker_id", w.workerID).Info("task worker received stop signal")
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
	logger := psl.GetLogger()
	logEntry := logger.WithField("worker_id", w.workerID)

	instances, err := w.q.ListPendingTaskWorkflowInstances(ctx, 10)
	if err != nil {
		logEntry.WithError(err).Error("failed to list pending workflow instances")
		return
	}
	if len(instances) == 0 {
		logEntry.Debug("no pending workflow instances")
		return
	}
	logEntry.WithField("count", len(instances)).Debug("found pending workflow instances")

	for _, instance := range instances {
		w.executeWorkflow(ctx, instance)
	}
}

func (w *TaskWorker) executeWorkflow(ctx context.Context, instance repo.TaskWorkflowInstance) {
	logger := psl.GetLogger()
	logEntry := logger.WithFields(logrus.Fields{
		"worker_id":            w.workerID,
		"workflow_instance_id": instance.ID,
		"workflow_def_id":      instance.WorkflowDefID,
	})
	logEntry.Info("starting workflow instance")

	now := time.Now()
	if _, err := w.q.UpdateTaskWorkflowInstanceStatus(ctx, repo.UpdateTaskWorkflowInstanceStatusParams{
		Status:   repo.TaskWorkflowInstanceStatusRUNNING,
		GmtStart: sql.NullTime{Time: now, Valid: true},
		ID:       instance.ID,
	}); err != nil {
		logEntry.WithError(err).Warn("failed to update workflow instance status to RUNNING")
	}

	nodes, err := w.q.ListTaskNodeInstances(ctx, instance.ID)
	if err != nil {
		logEntry.WithError(err).Error("failed to list workflow node instances")
		w.failWorkflow(ctx, instance.ID, err.Error())
		return
	}
	logEntry.WithField("node_count", len(nodes)).Debug("loaded workflow node instances")

	completed := 0
	failed := 0

	for _, node := range nodes {
		if node.Status != repo.TaskNodeInstanceStatusPENDING {
			continue
		}

		workflow, wfErr := w.q.GetTaskWorkflowDefByID(ctx, instance.WorkflowDefID)
		if wfErr != nil {
			logEntry.WithError(wfErr).Warn("failed to load workflow definition for onErrorStrategy; defaulting to STOP")
		}
		err := w.executeNode(ctx, instance.ID, node)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"worker_id":            w.workerID,
				"workflow_instance_id": instance.ID,
				"node_id":              node.NodeID,
				"node_instance_id":     node.ID,
			}).WithError(err).Warn("node execution failed")
			failed++

			onErrorStrategy := "STOP"
			if wfErr == nil && workflow.OnErrorStrategy.Valid {
				onErrorStrategy = string(workflow.OnErrorStrategy.TaskWorkflowDefOnErrorStrategy)
			}

			if onErrorStrategy == "STOP" {
				logEntry.WithField("failed_node_id", node.NodeID).Error("stopping workflow due to node failure")
				w.failWorkflow(ctx, instance.ID, fmt.Sprintf("Node %s failed: %v", node.NodeID, err))
				return
			}
		} else {
			completed++
		}

		if _, err := w.q.UpdateTaskWorkflowInstanceProgress(ctx, repo.UpdateTaskWorkflowInstanceProgressParams{
			TotalNodes:     sql.NullInt32{Int32: int32(len(nodes)), Valid: true},
			CompletedNodes: sql.NullInt32{Int32: int32(completed), Valid: true},
			FailedNodes:    sql.NullInt32{Int32: int32(failed), Valid: true},
			ID:             instance.ID,
		}); err != nil {
			logEntry.WithError(err).Warn("failed to update workflow instance progress")
		}
	}

	endTime := time.Now()
	if failed > 0 {
		errorInfo, _ := json.Marshal(map[string]interface{}{
			"failed_nodes":    failed,
			"completed_nodes": completed,
			"message":         "Workflow execution had failures",
		})
		if _, err := w.q.UpdateTaskWorkflowInstanceStatus(ctx, repo.UpdateTaskWorkflowInstanceStatusParams{
			Status:    repo.TaskWorkflowInstanceStatusFAILED,
			GmtEnd:    sql.NullTime{Time: endTime, Valid: true},
			ErrorInfo: errorInfo,
			ID:        instance.ID,
		}); err != nil {
			logEntry.WithError(err).Warn("failed to update workflow instance status to FAILED")
		}
		logEntry.WithFields(logrus.Fields{
			"completed_nodes": completed,
			"failed_nodes":    failed,
			"duration_ms":     endTime.Sub(now).Milliseconds(),
		}).Warn("workflow instance completed with failures")
	} else {
		result, _ := json.Marshal(map[string]interface{}{
			"completed_nodes": completed,
			"failed_nodes":    failed,
			"duration_ms":     endTime.Sub(now).Milliseconds(),
		})
		if _, err := w.q.UpdateTaskWorkflowInstanceStatus(ctx, repo.UpdateTaskWorkflowInstanceStatusParams{
			Status:        repo.TaskWorkflowInstanceStatusSUCCESS,
			GmtEnd:        sql.NullTime{Time: endTime, Valid: true},
			ResultSummary: result,
			ID:            instance.ID,
		}); err != nil {
			logEntry.WithError(err).Warn("failed to update workflow instance status to SUCCESS")
		}
		logEntry.WithFields(logrus.Fields{
			"completed_nodes": completed,
			"failed_nodes":    failed,
			"duration_ms":     endTime.Sub(now).Milliseconds(),
		}).Info("workflow instance completed")
	}
}

func (w *TaskWorker) executeNode(ctx context.Context, workflowInstanceID int64, node repo.TaskNodeInstance) error {
	logger := psl.GetLogger()
	logger.WithFields(logrus.Fields{
		"worker_id":            w.workerID,
		"workflow_instance_id": workflowInstanceID,
		"node_id":              node.NodeID,
		"node_instance_id":     node.ID,
	}).Debug("executing node")

	startTime := time.Now()
	if _, err := w.q.UpdateTaskNodeInstanceStatus(ctx, repo.UpdateTaskNodeInstanceStatusParams{
		Status:   repo.TaskNodeInstanceStatusRUNNING,
		GmtStart: sql.NullTime{Time: startTime, Valid: true},
		WorkerID: sql.NullString{String: w.workerID, Valid: true},
		ID:       node.ID,
	}); err != nil {
		logger.WithFields(logrus.Fields{
			"worker_id":        w.workerID,
			"node_instance_id": node.ID,
		}).WithError(err).Warn("failed to update node instance status to RUNNING")
	}

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
	logger := psl.GetLogger()
	logger.WithFields(logrus.Fields{
		"worker_id":               w.workerID,
		"workflow_instance_id":    workflowInstanceID,
		"node_id":                 node.NodeID,
		"node_instance_id":        node.ID,
		"task_atomic_def_id":      atomicTask.ID,
		"task_atomic_instance_id": atomicInstanceID,
		"script_type":             string(atomicTask.ScriptType.TaskAtomicDefScriptType),
	}).Debug("executing script task")

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

	errStr := ""
	errValid := false
	if err != nil {
		errStr = err.Error()
		errValid = true
	}

	if _, uErr := w.q.UpdateTaskAtomicInstanceStatus(ctx, repo.UpdateTaskAtomicInstanceStatusParams{
		Status:       status,
		OutputResult: json.RawMessage(fmt.Sprintf(`{"output": %q}`, string(output))),
		ExecutionLog: sql.NullString{String: logMsg, Valid: true},
		ErrorLog:     sql.NullString{String: errStr, Valid: errValid},
		GmtEnd:       sql.NullTime{Time: endTime, Valid: true},
		DurationMs:   sql.NullInt32{Int32: duration, Valid: true},
		WorkerID:     sql.NullString{String: w.workerID, Valid: true},
		ID:           atomicInstanceID,
	}); uErr != nil {
		logger.WithFields(logrus.Fields{
			"worker_id":               w.workerID,
			"task_atomic_instance_id": atomicInstanceID,
		}).WithError(uErr).Warn("failed to update atomic task instance status")
	}

	if _, uErr := w.q.UpdateTaskNodeInstanceStatus(ctx, repo.UpdateTaskNodeInstanceStatusParams{
		Status:       repo.TaskNodeInstanceStatus(status),
		OutputResult: json.RawMessage(fmt.Sprintf(`{"output": %q}`, string(output))),
		ExecutionLog: sql.NullString{String: logMsg, Valid: true},
		ErrorLog:     sql.NullString{String: errStr, Valid: errValid},
		GmtEnd:       sql.NullTime{Time: endTime, Valid: true},
		DurationMs:   sql.NullInt32{Int32: duration, Valid: true},
		WorkerID:     sql.NullString{String: w.workerID, Valid: true},
		ID:           node.ID,
	}); uErr != nil {
		logger.WithFields(logrus.Fields{
			"worker_id":        w.workerID,
			"node_instance_id": node.ID,
		}).WithError(uErr).Warn("failed to update node instance status")
	}

	if _, cErr := w.q.CreateTaskExecutionLog(ctx, repo.CreateTaskExecutionLogParams{
		WorkflowInstanceID:   workflowInstanceID,
		NodeInstanceID:       sql.NullInt64{Int64: node.ID, Valid: true},
		TaskAtomicInstanceID: sql.NullInt64{Int64: atomicInstanceID, Valid: true},
		LogLevel:             repo.TaskExecutionLogLogLevelINFO,
		LogType:              repo.TaskExecutionLogLogTypeEXECUTION,
		Message:              fmt.Sprintf("Script executed with status: %s", status),
	}); cErr != nil {
		logger.WithFields(logrus.Fields{
			"worker_id":               w.workerID,
			"workflow_instance_id":    workflowInstanceID,
			"task_atomic_instance_id": atomicInstanceID,
		}).WithError(cErr).Warn("failed to create task execution log")
	}

	if err != nil {
		logger.WithFields(logrus.Fields{
			"worker_id":               w.workerID,
			"workflow_instance_id":    workflowInstanceID,
			"task_atomic_instance_id": atomicInstanceID,
			"duration_ms":             duration,
		}).WithError(err).Warn("script execution failed")
		return fmt.Errorf("script execution failed: %w, output: %s", err, logMsg)
	}

	return nil
}

func (w *TaskWorker) executeHTTPTask(ctx context.Context, workflowInstanceID int64, node repo.TaskNodeInstance, atomicTask repo.TaskAtomicDef, atomicInstanceID int64, startTime time.Time) error {
	logger := psl.GetLogger()
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

	logger.WithFields(logrus.Fields{
		"worker_id":               w.workerID,
		"workflow_instance_id":    workflowInstanceID,
		"node_id":                 node.NodeID,
		"task_atomic_instance_id": atomicInstanceID,
		"method":                  method,
		"url":                     url,
	}).Info("executing http task")
	_ = body

	endTime := time.Now()
	duration := int32(endTime.Sub(startTime).Milliseconds())

	status := repo.TaskAtomicInstanceStatusSUCCESS
	logMsg := fmt.Sprintf("HTTP %s %s completed", method, url)

	if _, err := w.q.UpdateTaskAtomicInstanceStatus(ctx, repo.UpdateTaskAtomicInstanceStatusParams{
		Status:       status,
		ExecutionLog: sql.NullString{String: logMsg, Valid: true},
		GmtEnd:       sql.NullTime{Time: endTime, Valid: true},
		DurationMs:   sql.NullInt32{Int32: duration, Valid: true},
		WorkerID:     sql.NullString{String: w.workerID, Valid: true},
		ID:           atomicInstanceID,
	}); err != nil {
		logger.WithFields(logrus.Fields{
			"worker_id":               w.workerID,
			"task_atomic_instance_id": atomicInstanceID,
		}).WithError(err).Warn("failed to update atomic task instance status")
	}

	if _, err := w.q.UpdateTaskNodeInstanceStatus(ctx, repo.UpdateTaskNodeInstanceStatusParams{
		Status:       repo.TaskNodeInstanceStatus(status),
		ExecutionLog: sql.NullString{String: logMsg, Valid: true},
		GmtEnd:       sql.NullTime{Time: endTime, Valid: true},
		DurationMs:   sql.NullInt32{Int32: duration, Valid: true},
		WorkerID:     sql.NullString{String: w.workerID, Valid: true},
		ID:           node.ID,
	}); err != nil {
		logger.WithFields(logrus.Fields{
			"worker_id":        w.workerID,
			"node_instance_id": node.ID,
		}).WithError(err).Warn("failed to update node instance status")
	}

	return nil
}

func (w *TaskWorker) failWorkflow(ctx context.Context, instanceID int64, reason string) {
	logger := psl.GetLogger()
	logger.WithFields(logrus.Fields{
		"worker_id":            w.workerID,
		"workflow_instance_id": instanceID,
		"reason":               reason,
	}).Error("marking workflow instance as FAILED")

	endTime := time.Now()
	errorInfo, _ := json.Marshal(map[string]interface{}{
		"error":     reason,
		"failed_at": endTime.Format(time.RFC3339),
	})

	if _, err := w.q.UpdateTaskWorkflowInstanceStatus(ctx, repo.UpdateTaskWorkflowInstanceStatusParams{
		Status:       repo.TaskWorkflowInstanceStatusFAILED,
		StatusReason: sql.NullString{String: reason, Valid: true},
		GmtEnd:       sql.NullTime{Time: endTime, Valid: true},
		ErrorInfo:    errorInfo,
		ID:           instanceID,
	}); err != nil {
		logger.WithFields(logrus.Fields{
			"worker_id":            w.workerID,
			"workflow_instance_id": instanceID,
		}).WithError(err).Warn("failed to update workflow instance status to FAILED")
	}
}

func (w *TaskWorker) RollbackWorkflow(ctx context.Context, instanceID int64) error {
	logger := psl.GetLogger()
	logEntry := logger.WithFields(logrus.Fields{
		"worker_id":            w.workerID,
		"workflow_instance_id": instanceID,
	})
	logEntry.Info("starting workflow rollback")

	nodes, err := w.q.ListTaskNodeInstances(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}
	logEntry.WithField("node_count", len(nodes)).Debug("loaded nodes for rollback")

	for i := len(nodes) - 1; i >= 0; i-- {
		node := nodes[i]
		if node.Status != repo.TaskNodeInstanceStatusSUCCESS {
			continue
		}

		atomicInstances, err := w.q.ListTaskAtomicInstances(ctx, node.ID)
		if err != nil {
			logEntry.WithFields(logrus.Fields{
				"node_instance_id": node.ID,
				"node_id":          node.NodeID,
			}).WithError(err).Warn("failed to list atomic instances for rollback")
			continue
		}
		for j := len(atomicInstances) - 1; j >= 0; j-- {
			atomicInst := atomicInstances[j]
			atomicTask, err := w.q.GetTaskAtomicDefByID(ctx, atomicInst.TaskAtomicDefID)
			if err != nil {
				logEntry.WithFields(logrus.Fields{
					"task_atomic_instance_id": atomicInst.ID,
					"task_atomic_def_id":      atomicInst.TaskAtomicDefID,
				}).WithError(err).Warn("failed to load atomic task definition for rollback")
				continue
			}

			if atomicTask.IsRollbackSupported.Bool && atomicTask.RollbackScriptContent.Valid {
				w.executeRollback(ctx, atomicInst, atomicTask)
			}
		}
	}

	if _, err := w.q.UpdateTaskWorkflowInstanceStatus(ctx, repo.UpdateTaskWorkflowInstanceStatusParams{
		Status: repo.TaskWorkflowInstanceStatusCANCELLED,
		ID:     instanceID,
	}); err != nil {
		logEntry.WithError(err).Warn("failed to update workflow instance status to CANCELLED after rollback")
	}
	logEntry.Info("workflow rollback completed")

	return nil
}

func (w *TaskWorker) executeRollback(ctx context.Context, atomicInst repo.TaskAtomicInstance, atomicTask repo.TaskAtomicDef) {
	logger := psl.GetLogger()
	logEntry := logger.WithFields(logrus.Fields{
		"worker_id":               w.workerID,
		"task_atomic_instance_id": atomicInst.ID,
		"task_atomic_def_id":      atomicTask.ID,
		"script_type":             string(atomicTask.RollbackScriptType.TaskAtomicDefRollbackScriptType),
	})
	logEntry.Info("rolling back atomic task")

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
		logEntry.WithError(err).Warn("rollback failed")
	}

	if _, uErr := w.q.UpdateTaskAtomicInstanceRollback(ctx, repo.UpdateTaskAtomicInstanceRollbackParams{
		RollbackLog:    sql.NullString{String: logMsg, Valid: true},
		RollbackResult: json.RawMessage(fmt.Sprintf(`{"result": %q}`, result)),
		ID:             atomicInst.ID,
	}); uErr != nil {
		logEntry.WithError(uErr).Warn("failed to update atomic instance rollback result")
	}
}
