# Allora Monitor Frontend

A modern web interface for the Allora Monitor backend, built with Next.js and Tailwind CSS.

## Features

-   Dashboard with system status, database stats, and active topics
-   Competitions listing and details
-   Topics exploration with inference data
-   Detailed statistics and monitoring information
-   Responsive design for desktop and mobile devices
-   Dark mode support

## Getting Started

### Prerequisites

-   Node.js 18.0.0 or later
-   Backend server running (see [backend README](../backend/README.md))

### Installation

1. Clone the repository:

```bash
git clone https://github.com/sonamu/allora-monitor.git
cd allora-monitor/frontend
```

2. Install dependencies:

```bash
npm install
# or
yarn install
# or
bun install
```

3. Configure the backend URL:

Edit `.env.local` file to configure the application:

```
# Backend API URL
NEXT_PUBLIC_API_URL=http://localhost:8080

# Set to 'true' to use mock data when backend is not available
NEXT_PUBLIC_USE_MOCK_DATA=false

# Enable debug logging
NEXT_PUBLIC_DEBUG=false
```

### Development

Run the development server:

```bash
npm run dev
# or
yarn dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

### Building for Production

Build the application for production:

```bash
npm run build
# or
yarn build
# or
bun build
```

Start the production server:

```bash
npm run start
# or
yarn start
# or
bun start
```

## Troubleshooting Connection Issues

If you encounter connection issues between the frontend and backend:

1. **Verify Backend Server**: Make sure the backend server is running on the correct port (default: 8080)

2. **Check CORS Configuration**: The backend needs to allow cross-origin requests from the frontend. The backend should have CORS headers properly configured.

3. **Use Mock Data**: If you need to develop the frontend without a running backend, you can enable mock data by setting `NEXT_PUBLIC_USE_MOCK_DATA=true` in your `.env.local` file.

4. **Network Issues**: If running in different environments (e.g., Docker, different networks), make sure the frontend can reach the backend server.

5. **API URL Configuration**: Ensure the `NEXT_PUBLIC_API_URL` in `.env.local` points to the correct backend URL.

6. **Debug Mode**: Enable debug logging by setting `NEXT_PUBLIC_DEBUG=true` to see detailed API request information in the browser console.

## Project Structure

```
src/
├── app/                  # Next.js app router pages
│   ├── competitions/     # Competitions pages
│   ├── topics/           # Topics pages
│   ├── stats/            # Statistics page
│   ├── layout.tsx        # Root layout with navigation
│   └── page.tsx          # Dashboard page
├── components/           # Reusable UI components
│   └── ui/               # Basic UI components
├── lib/                  # Utility functions and API clients
│   ├── api.ts            # Backend API client
│   └── utils.ts          # Helper functions
└── ...
```

## Technologies Used

-   [Next.js](https://nextjs.org/) - React framework
-   [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
-   [SWR](https://swr.vercel.app/) - React Hooks for data fetching
-   [Axios](https://axios-http.com/) - Promise based HTTP client
-   [date-fns](https://date-fns.org/) - Date utility library
-   [Recharts](https://recharts.org/) - Charting library

## License

This project is licensed under the MIT License - see the LICENSE file for details.
