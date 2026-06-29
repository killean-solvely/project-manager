import { Routes, Route, Navigate } from 'react-router-dom'
import { PortfolioPage } from './pages/PortfolioPage'
import { ProjectPage } from './pages/ProjectPage'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<PortfolioPage />} />
      <Route path="/projects/:id" element={<ProjectPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
