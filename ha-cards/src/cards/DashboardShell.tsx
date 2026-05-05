import type { ReactNode } from 'react';

interface DashboardShellProps {
  topBar: ReactNode;
  detail: ReactNode;
  mode?: 'standalone' | 'embedded';
}

export function DashboardShell({ topBar, detail, mode = 'standalone' }: DashboardShellProps) {
  return (
    <div className={`ajaxbridge-theme app-shell ${mode === 'embedded' ? 'app-shell--embedded' : ''}`}>
      <div className="app-shell__halo app-shell__halo--cyan" />
      <div className="app-shell__halo app-shell__halo--red" />
      <div className={`dashboard-frame ${mode === 'embedded' ? 'dashboard-frame--embedded' : ''}`}>
        {topBar}
        {detail}
      </div>
    </div>
  );
}
