import { useMemo, useState } from 'react';
import { DragDropContext, Draggable, Droppable, type DropResult } from 'react-beautiful-dnd';
import { Plus } from 'lucide-react';

import { useMoveTask, useCreateTask } from '@controllers/hooks/useTasks';
import { Button } from '@views/components/ui/button';
import { Badge } from '@views/components/ui/badge';
import { TaskCard } from '@views/components/tasks/TaskCard';
import { TaskModal } from '@views/components/tasks/TaskModal';
import { CreateTaskDialog } from '@views/components/tasks/CreateTaskDialog';
import { TASK_STATUSES, TASK_STATUS_ORDER, STATUS_COLUMN_COLORS } from '@utils/constants';
import type { Task, TaskStatus } from '@models/types/task.types';

interface Props {
  projectId: string;
  groupId: string;
  tasks: Task[];
}

/**
 * Kanban-доска с 4 колонками: todo / in_progress / done / blocked.
 *
 * Поток drag-and-drop:
 *  1. react-beautiful-dnd сообщает результат drop в onDragEnd
 *  2. Считаем новый статус (по destination.droppableId) и порядок
 *     (по destination.index среди задач этой колонки)
 *  3. Дёргаем useMoveTask — мутация делает оптимистичный апдейт
 *     кэша TanStack Query и шлёт запрос в /tasks/move
 *  4. При ошибке состояние откатится автоматически
 */
export function KanbanBoard({ projectId, groupId, tasks }: Props) {
  const moveTask = useMoveTask(projectId);
  const createTask = useCreateTask(projectId);

  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [createInColumn, setCreateInColumn] = useState<TaskStatus | null>(null);

  // Группируем по статусам и сортируем по полю order
  const grouped = useMemo(() => {
    const result: Record<TaskStatus, Task[]> = {
      todo: [],
      in_progress: [],
      done: [],
      blocked: [],
    };
    for (const t of tasks) result[t.status].push(t);
    for (const key of Object.keys(result) as TaskStatus[]) {
      result[key].sort((a, b) => a.order - b.order);
    }
    return result;
  }, [tasks]);

  const handleDragEnd = (result: DropResult) => {
    const { destination, source, draggableId } = result;
    if (!destination) return;

    // Дроп в то же место
    if (
      destination.droppableId === source.droppableId &&
      destination.index === source.index
    ) {
      return;
    }

    moveTask.mutate({
      taskId: draggableId,
      newStatus: destination.droppableId as TaskStatus,
      newOrder: destination.index,
    });
  };

  return (
    <>
      <DragDropContext onDragEnd={handleDragEnd}>
        <div className="flex gap-4 overflow-x-auto pb-4">
          {TASK_STATUS_ORDER.map((statusKey) => {
            const status = statusKey as TaskStatus;
            const items = grouped[status];

            return (
              <Droppable droppableId={status} key={status}>
                {(provided, snapshot) => (
                  <div
                    ref={provided.innerRef}
                    {...provided.droppableProps}
                    className={`kanban-column ${
                      snapshot.isDraggingOver ? 'bg-muted' : ''
                    }`}
                  >
                    {/* Заголовок колонки */}
                    <div className="mb-3 flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span
                          className="h-2.5 w-2.5 rounded-full"
                          style={{ backgroundColor: STATUS_COLUMN_COLORS[status] }}
                        />
                        <h3 className="text-sm font-semibold">{TASK_STATUSES[status]}</h3>
                        <Badge variant="secondary" className="h-5 px-1.5 text-xs">
                          {items.length}
                        </Badge>
                      </div>
                      <Button
                        size="icon"
                        variant="ghost"
                        className="h-7 w-7"
                        onClick={() => setCreateInColumn(status)}
                        aria-label="Добавить задачу"
                      >
                        <Plus className="h-4 w-4" />
                      </Button>
                    </div>

                    {/* Задачи */}
                    <div className="min-h-[100px] flex-1">
                      {items.map((task, idx) => (
                        <Draggable key={task.id} draggableId={task.id} index={idx}>
                          {(prov, snap) => (
                            <div
                              ref={prov.innerRef}
                              {...prov.draggableProps}
                              {...prov.dragHandleProps}
                            >
                              <TaskCard
                                task={task}
                                onClick={() => setSelectedTask(task)}
                                isDragging={snap.isDragging}
                              />
                            </div>
                          )}
                        </Draggable>
                      ))}
                      {provided.placeholder}

                      {items.length === 0 && (
                        <p className="rounded-md border border-dashed p-4 text-center text-xs text-muted-foreground">
                          Пусто
                        </p>
                      )}
                    </div>
                  </div>
                )}
              </Droppable>
            );
          })}
        </div>
      </DragDropContext>

      {/* Модалка задачи */}
      <TaskModal
        task={selectedTask}
        projectId={projectId}
        groupId={groupId}
        open={!!selectedTask}
        onOpenChange={(o) => !o && setSelectedTask(null)}
      />

      {/* Создание задачи */}
      <CreateTaskDialog
        open={!!createInColumn}
        onOpenChange={(o) => !o && setCreateInColumn(null)}
        loading={createTask.isPending}
        onSubmit={(values) => {
          createTask.mutate(
            {
              title: values.title,
              description: values.description,
              priority: values.priority,
              projectId,
              dueDate: values.dueDate,
              assigneeId: values.assigneeId,
            },
            {
              onSuccess: () => setCreateInColumn(null),
            },
          );
        }}
      />
    </>
  );
}
