# Ragtime Frontend

Next.js frontend for the Ragtime versioned RAG knowledge base system.

## Tech Stack

- **Next.js 16.1.6** - React framework with App Router
- **React 19** - UI library
- **TypeScript 5** - Type safety
- **Tailwind CSS 4** - Utility-first CSS framework
- **ESLint** - Code linting

## Getting Started

### Prerequisites

- Node.js 20+ recommended
- Backend API running (default: `http://localhost:8080`)

### Installation

```bash
npm install
```

### Environment Setup

1. Copy the environment template:
```bash
cp .env.example .env.local
```

2. Update `.env.local` with your backend API URL if different from default and set Clerk keys:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_API_VERSION=v1

NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_your_key
CLERK_SECRET_KEY=sk_test_your_key
CLERK_WEBHOOK_SECRET=whsec_your_webhook_secret
```

### Development

Run the development server:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser.

### Building for Production

```bash
npm run build
npm run start
```

### Linting

```bash
npm run lint
```

## Project Structure

```
frontend/
├── app/                 # Next.js App Router pages
│   ├── layout.tsx      # Root layout
│   └── page.tsx        # Home page
├── public/             # Static assets
├── .env.example        # Environment variables template
├── next.config.ts      # Next.js configuration
├── tailwind.config.ts  # Tailwind CSS configuration
├── tsconfig.json       # TypeScript configuration
└── eslint.config.mjs   # ESLint configuration
```

## Backend API Integration

The backend API is expected to be running at the URL specified in `NEXT_PUBLIC_API_URL`.

### Key Endpoints

- `GET /v1/kb` - List knowledge bases
- `POST /v1/kb` - Create knowledge base
- `POST /v1/kb/{kb_id}/documents` - Upload document
- `POST /v1/kb/{kb_id}/query` - Query knowledge base

See the backend documentation for full API reference.

## Development Guidelines

- Follow the repository pattern established in the backend
- Keep components focused and reusable
- Use TypeScript strictly (avoid `any` types)
- Follow Tailwind CSS utility-first approach
- Ensure all components are accessible (a11y)
- Write meaningful commit messages

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [TypeScript Documentation](https://www.typescriptlang.org/docs)
