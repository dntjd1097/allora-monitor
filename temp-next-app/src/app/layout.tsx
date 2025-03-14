import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
    title: 'Allora Monitor',
    description: 'Monitoring dashboard for Allora Network',
};

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    const currentYear = new Date().getFullYear();

    return (
        <html lang="en">
            <body className="bg-gray-50">
                <div className="flex flex-col min-h-screen">
                    <Header />
                    <main className="flex-1 container mx-auto px-4 py-6">
                        {children}
                    </main>
                    <Footer year={currentYear} />
                </div>
            </body>
        </html>
    );
}

function Header() {
    return (
        <header className="bg-gradient-to-r from-indigo-600 to-purple-600 text-white shadow-md">
            <div className="container mx-auto px-4">
                <div className="flex justify-between items-center py-4">
                    <div className="flex items-center space-x-2">
                        <h1 className="text-2xl font-bold">
                            Allora Monitor
                        </h1>
                    </div>
                </div>
            </div>
        </header>
    );
}

function Footer({ year }: { year: number }) {
    return (
        <footer className="bg-gray-100 border-t border-gray-200 py-4 mt-8">
            <div className="container mx-auto px-4 text-center text-[var(--foreground)]">
                <p>
                    © {year} Allora Monitor. All rights
                    reserved.
                </p>
            </div>
        </footer>
    );
}
