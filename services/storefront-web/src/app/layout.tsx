import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import Navbar from "@/components/layout/Navbar";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "GO commerce™",
  description: "Enterprise Microservice E-commerce Platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${inter.className} min-h-screen bg-gray-50 flex flex-col`}>
        <Navbar />
        <main className="flex-1 pb-16">
          {children}
        </main>
        <footer className="border-t bg-white py-6 mt-auto">
          <div className="container text-center text-sm text-gray-500">
            &copy; {new Date().getFullYear()} GO commerce&trade;. All rights reserved.
          </div>
        </footer>
      </body>
    </html>
  );
}
