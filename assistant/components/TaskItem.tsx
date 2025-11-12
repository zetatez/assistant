
interface Task {
  id: number;
  title: string;
  description: string;
}

export default function TaskItem({ task }: { task: Task }) {
  return (
    <div>
      <h2 className="text-lg font-semibold">{task.title}</h2>
      <p className="text-sm text-gray-600">{task.description}</p>
    </div>
  );
}
