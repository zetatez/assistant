
"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import api from "@/lib/api";
import { useStore } from "@/lib/store";
import TaskItem from "@/components/TaskItem";
import TaskForm from "@/components/TaskForm";

export default function TasksPage() {
  const { tasks, setTasks, removeTask } = useStore();
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    async function fetchTasks() {
      setLoading(true);
      try {
        const res = await api.get("/tasks");
        setTasks(res.data);
      } catch (err) {
        alert("获取任务失败");
      } finally {
        setLoading(false);
      }
    }
    fetchTasks();
  }, [setTasks]);

  async function handleDelete(id: number) {
    await api.delete(`/tasks/${id}`);
    removeTask(id);
  }

  return (
    <div className="p-6 max-w-3xl mx-auto space-y-6">
      <h1 className="text-3xl font-bold mb-6">任务管理</h1>
      <TaskForm />

      {loading && <p>加载中...</p>}
      {!loading && tasks.length === 0 && <p>暂无任务</p>}

      <div className="space-y-3">
        {tasks.map((task) => (
          <Card key={task.id} className="p-4 flex justify-between items-center">
            <TaskItem task={task} />
            <Button
              variant="destructive"
              onClick={() => handleDelete(task.id)}
              size="sm"
            >
              删除
            </Button>
          </Card>
        ))}
      </div>
    </div>
  );
}
