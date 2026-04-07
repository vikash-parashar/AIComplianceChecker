# 🛡️ AI Compliance Checker (SaaS)

**A high-performance, intelligent SaaS platform designed to automatically validate SMS marketing, Privacy Policies, and Application Configurations against strict regulatory standards (HIPAA, GDPR, CCPA, A2P 10DLC).**

Built for enterprise-grade scalability, this platform combines the raw speed of a **Golang** backend with the deep semantic reasoning of **OpenAI (GPT-4o)**, all wrapped in a premium, glassmorphism **Next.js** user interface.

---

## 🏗️ Architecture & Tech Stack

This project is divided into an isolated API backend and a responsive frontend portal, connected securely via JWT authentication and Stripe billing.

### ⚡ Backend (Core Engine)
*   **Golang (Gin Framework)**: Lightning-fast concurrent API that scales elegantly.
*   **PostgreSQL**: Secure, relational data storage tracking user credits, compliance violation histories, and mock authentication hashes.
*   **OpenAI API (GPT-4o)**: Powers the compliance engine. We utilize **advanced engineered system prompts** that strictly enforce complex data privacy logic onto user inputs.
*   **Stripe SDK**: Integrated Checkout session modeling for localized and global credit purchasing (Cards, Apple Pay, Google Pay, UPI).
*   **Asynchronous Notifications**: Utilizes Go Routines to fire off Email (`net/smtp`) and SMS (Twilio HTTP Hooks) lifecycle events without blocking API responses.

### 🖥️ Frontend (Client Portal)
*   **Next.js (App Router)**: Modern React framework optimized with Turbopack.
*   **Vanilla CSS Glassmorphism**: Tailored, premium UI design featuring dynamic tabs, smooth hovering mechanics, and custom frosted-glass interfaces (No Tailwind overhead).
*   **Client-Side Analytics**: Asynchronously tracks user usage credits, updating state dynamically post-scan, providing real-time "Low Credit" warnings before limits are exhausted.

### 🐳 DevOps
*   **Multi-Stage Dockerfiles**: The Go backend compiles down to a raw Alpine binary, and Next.js is configured for `standalone` server deployment, drastically minimizing cloud footprint.

---

## ⚙️ How to Run Locally

### Prerequisites
1. Docker & Docker Compose
2. Node.js (v18+)
3. Go (v1.21+)
4. An OpenAI API Key (Stripe / Mail keys are optional; the app features built-in mocks).

### Step 1: Spin up the Database
From the root directory, launch the Postgres container:
```bash
docker-compose up -d
```
*(This bridges Postgres to `localhost:5433`)*

### Step 2: Start the Go Backend
Open a new terminal session in `/backend`:
```bash
cd backend
export OPENAI_API_KEY="sk-proj-YOUR_REAL_KEY"
export DATABASE_URL="postgres://ai_compliance:password123@localhost:5433/aicompliance_db?sslmode=disable"
go run main.go
```
*(The backend runs on `http://localhost:8085`)*

### Step 3: Start the Next.js Frontend
Open a new terminal session in `/frontend`:
```bash
cd frontend
npm install
npm run dev
```
*(The frontend runs on `http://localhost:3000`)*

---

## 💼 Go-To-Market & Selling Strategy

The AI Compliance Checker is a highly niche, B2B (Business-to-Business) SaaS tool. Businesses face massive fines for non-compliance (A2P 10DLC violations block marketing funnels, GDPR violations result in million-dollar suits). **Your platform is an insurance policy for them.**

### Target Audience:
1.  **Marketing Agencies:** Agencies sending mass SMS blasts need to ensure their copy includes strict `STOP/HELP` logic and avoids SHAFT terminology so they don't get blocked by AT&T/T-Mobile.
2.  **HealthTech Startups:** Smaller startups dealing with patient data who cannot afford $500/hr lawyers to review their app's logging parameters or JSON configs.
3.  **Indie Developers:** Solopreneurs building apps that need to quickly throw up a GDPR/CCPA compliant privacy policy but don't know the nuances of "Right to be Forgotten".

### Pricing Model (The "Credit" System):
*   **Freemium Hook:** Offer 10 Free Credits upon logging in. This allows them to test the AI's power immediately and builds trust.
*   **Pay-as-you-go:** Since OpenAI charges per token, a subscription might be risky early on. Charge a flat $20.00 for 100 Scans. The Stripe integration is already scoped for this!

### Marketing Channels (How to get users):
*   **Product Hunt Launch:** Position the SaaS as "The Grammarly for Regulatory Compliance." Showcase the slick UI.
*   **Cold Emailing Agencies:** Scrape lists of digital marketing SMS agencies. Send them a cold email stating: *"Carrier blocking algorithms changed in 2024. Your clients' SMS campaigns might get killed. Run your text copy through our AI before broadcasting to guarantee 10DLC compliance."*
*   **Open Source "Widgets":** Expose a paid API endpoint so developers can integrate your checker directly into their CMS or VSCode IDE.

### Highlighting Features to Investors/Buyers
If you ever want to sell this SaaS to a private equity firm via platforms like **Acquire.com**, highlight the fact that:
1.  **Architecture is Enterprise-Ready:** You didn't use a no-code tool. It’s built on scalable Go, Docker, and PostgreSQL.
2.  **Monetization is Hardcoded:** The Stripe logic and Webhook security stops credit theft natively.
3.  **Low Overhead:** The multi-stage Dockerfiles keep hosting costs under $10/month while securely supporting thousands of concurrent users.
