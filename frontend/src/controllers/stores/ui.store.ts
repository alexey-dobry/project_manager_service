import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

/**
 * Глобальное UI-состояние: тема, видимость сайдбара и т.п.
 */
type Theme = 'light' | 'dark' | 'system';

interface UIState {
  theme: Theme;
  isSidebarOpen: boolean;

  setTheme: (theme: Theme) => void;
  toggleSidebar: () => void;
  setSidebar: (open: boolean) => void;
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      theme: 'system',
      isSidebarOpen: true,

      setTheme: (theme) => {
        set({ theme });
        applyTheme(theme);
      },

      toggleSidebar: () => set((s) => ({ isSidebarOpen: !s.isSidebarOpen })),
      setSidebar: (open) => set({ isSidebarOpen: open }),
    }),
    {
      name: 'studentpm-ui-store',
      storage: createJSONStorage(() => localStorage),
      onRehydrateStorage: () => (state) => {
        if (state) applyTheme(state.theme);
      },
    },
  ),
);

function applyTheme(theme: Theme): void {
  const root = document.documentElement;
  root.classList.remove('light', 'dark');

  if (theme === 'system') {
    const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    root.classList.add(isDark ? 'dark' : 'light');
  } else {
    root.classList.add(theme);
  }
}
