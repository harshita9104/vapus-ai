'use client';

import { useEffect } from 'react';

export default function DocsPage() {
  useEffect(() => {
    // Redirect to the static documentation
    window.location.href = '/docs-static/';
  }, []);

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <h1 className="text-2xl font-bold mb-4">Loading Documentation...</h1>
        <p className="text-gray-600">
          Redirecting to VapusAI Documentation
        </p>
      </div>
    </div>
  );
}
