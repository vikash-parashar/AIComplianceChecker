import type { Metadata } from 'next'
import './globals.css'
import Link from 'next/link'

export const metadata: Metadata = {
  title: 'AI Compliance Checker',
  description: 'Validate SMS, Policies, and Configurations against HIPAA, GDPR, and A2P 10DLC.',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>
        <nav className="navbar">
          <Link href="/" className="logo">
            <span className="logo-accent">AI</span> Compliance
          </Link>
          <div className="nav-links">
            <Link href="/">Dashboard</Link>
            <Link href="/auth">Sign In</Link>
          </div>
        </nav>
        <main className="main-content">
          {children}
        </main>
      </body>
    </html>
  )
}
