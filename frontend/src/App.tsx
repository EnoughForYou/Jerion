import { useState } from 'react';
import { RouterProvider } from 'react-router-dom';
import { PreflightView } from './features/preflight/PreflightView';
import { router } from './app/routes';
import './style.css';
import './i18n';

export function App() {
  const [preflightPassed, setPreflightPassed] = useState(false);

  if (!preflightPassed) {
    return <PreflightView onComplete={() => setPreflightPassed(true)} />;
  }

  return <RouterProvider router={router} />;
}

export default App;
