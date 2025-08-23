'use client';

import { useEffect } from 'react';
import { useParams } from 'next/navigation';

export default function DocsStaticPathPage() {
  const params = useParams();
  
  useEffect(() => {
    // Get the path from params
    const path = Array.isArray(params.path) ? params.path.join('/') : params.path || '';
    
    // Redirect to the actual static documentation files
    window.location.href = `/docs-build/${path}`;
  }, [params]);

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
