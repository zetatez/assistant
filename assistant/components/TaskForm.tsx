
"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import api from "@/lib/api";
import { useStore } from "@/lib/store";

export default function TaskForm() {
  const [title, setTitle] = useState("");
  const [desc, setDesc] = useState("");
  const [loading, setLoading] = useState(false);
  const addTask = useStore((s) => s.addTask);

  async function handleAdd() {
    if (!title.trim()) return alert("请输入任务标题");
    setLoading(true);
    try {
      const res = await api.post("/tasks", { title, description: desc });
      addTask(res.data);
      setTitle("");
      setDesc("");
    } catch (err) {
      alert("添加任务失败");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="border p-4 rounded-xl shadow-sm space-y-3">
      <Input
        placeholder="任务标题"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
      />
      <Textarea
        placeholder="任务描述"
        value={desc}
        onChange={(e) => setDesc(e.target.value)}
      />
      <Button onClick={handleAdd} disabled={loading}>
        {loading ? "添加中..." : "添加任务"}
      </Button>
    </div>
  );
}
