import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Header } from './Header';

/**
 * Двухколоночный лейаут:
 * - левая колонка: фиксированный сайдбар
 * - правая: header + основное содержимое (Outlet от react-router)
 *
 * Подсказка по адаптиву: на мобиле сайдбар сворачивается в иконки
 * через Header → toggleSidebar; полноценный drawer оставлен на расширение.
 */
export function Layout() {
  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <Header />
        <main className="flex-1 p-4 md:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
