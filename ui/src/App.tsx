import { Route, Routes, Link } from 'react-router-dom'
import AboutPage from './pages/AboutPage'
import SettingsPage from './pages/SettingsPage'
import ContextPage from './pages/ContextPage'
import JobsPage from './pages/JobsPage'
import JobDetailsPage from './pages/JobDetailsPage'
import MemoryPage from './pages/MemoryPage'
import RepoPage from './pages/RepoPage'
import RepoDetailsPage from './pages/RepoDetailsPage'
import ProviderPage from './pages/ProviderPage'

export default function App() {
  return (
    <div className="min-h-screen">
      <nav className="border-b border-gray-800 px-6 py-3">
        <div className="max-w-5xl mx-auto flex items-center gap-6">
          <Link to="/" className="font-semibold text-white tracking-wide hover:text-gray-300">GhostConductor</Link>
          <Link to="/jobs" className="text-sm text-gray-400 hover:text-white">Summonings</Link>
          <Link to="/repos" className="text-sm text-gray-400 hover:text-white">Repos</Link>
          <Link to="/memory" className="text-sm text-gray-400 hover:text-white">Memory</Link>
          <Link to="/context" className="text-sm text-gray-400 hover:text-white">Context</Link>
          <Link to="/providers" className="text-sm text-gray-400 hover:text-white">Providers</Link>
          <Link to="/settings" className="text-sm text-gray-400 hover:text-white">Settings</Link>
        </div>
      </nav>
      <Routes>
        <Route path="/" element={<AboutPage />} />
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="/jobs" element={<JobsPage />} />
        <Route path="/jobs/:jobId" element={<JobDetailsPage />} />
        <Route path="/repos" element={<RepoPage />} />
        <Route path="/repos/:repoId" element={<RepoDetailsPage />} />
        <Route path="/repos/:repoId/jobs/:jobId" element={<RepoDetailsPage />} />
        <Route path="/memory" element={<MemoryPage />} />
        <Route path="/context" element={<ContextPage />} />
        <Route path="/providers" element={<ProviderPage />} />
      </Routes>
    </div>
  )
}
