import { Routes, Route, Navigate } from 'react-router-dom'
import { Navbar } from './components/layout/Navbar'
import { Footer } from './components/layout/Footer'
import { ProtectedRoute } from './components/layout/ProtectedRoute'

import { LandingPage } from './pages/LandingPage'
import { SignupPage } from './pages/SignupPage'
import { LoginPage } from './pages/LoginPage'
import { DashboardPage } from './pages/DashboardPage'
import { CreatePollPage } from './pages/CreatePollPage'
import { PollVotePage } from './pages/PollVotePage'
import { LiveResultsPage } from './pages/LiveResultsPage'
import { PollManagePage } from './pages/PollManagePage'

import './App.css'

export default function App() {
  return (
    <div className="app-shell">
      <Navbar />
      <main className="app-main-content">
        <Routes>
          {/* 1. Landing page */}
          <Route path="/" element={<LandingPage />} />

          {/* 2. Signup page */}
          <Route path="/signup" element={<SignupPage />} />

          {/* 3. Login page */}
          <Route path="/login" element={<LoginPage />} />

          {/* 4. Creator dashboard (Protected) */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <DashboardPage />
              </ProtectedRoute>
            }
          />

          {/* 5. Create poll (Protected) */}
          <Route
            path="/create"
            element={
              <ProtectedRoute>
                <CreatePollPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/create-poll"
            element={
              <ProtectedRoute>
                <CreatePollPage />
              </ProtectedRoute>
            }
          />

          {/* 6. Poll voting page (Public Audience) */}
          <Route path="/polls/:id" element={<PollVotePage />} />
          <Route path="/polls/:id/vote" element={<PollVotePage />} />
          <Route path="/poll/:id" element={<PollVotePage />} />
          <Route path="/vote/:id" element={<PollVotePage />} />

          {/* 7. Live results page (Public Presenter / Stream) */}
          <Route path="/polls/:id/results" element={<LiveResultsPage />} />
          <Route path="/poll/:id/results" element={<LiveResultsPage />} />
          <Route path="/results/:id" element={<LiveResultsPage />} />

          {/* 8. Poll management / details page (Protected) */}
          <Route
            path="/polls/:id/manage"
            element={
              <ProtectedRoute>
                <PollManagePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/manage/:id"
            element={
              <ProtectedRoute>
                <PollManagePage />
              </ProtectedRoute>
            }
          />

          {/* Fallback */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
      <Footer />
    </div>
  )
}
