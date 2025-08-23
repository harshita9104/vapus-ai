'use client';

import { useEffect, useRef } from 'react';
import { useParams } from 'next/navigation';

export default function DocsSubPage() {
  const iframeRef = useRef(null);
  const params = useParams();
  
  // Reconstruct the full path from the catch-all route
  const docsPath = params.slug ? `/${params.slug.join('/')}` : '/';

  useEffect(() => {
    // Update iframe src when path changes
    if (iframeRef.current) {
      const baseUrl = process.env.NODE_ENV === 'development' 
        ? 'http://localhost:3001' 
        : '/docs-static';
      iframeRef.current.src = `${baseUrl}${docsPath}`;
    }
  }, [docsPath]);

  return (
    <div className="docs-container" style={{ height: '100vh', width: '100%' }}>
      <iframe
        ref={iframeRef}
        style={{
          width: '100%',
          height: '100%',
          border: 'none',
        }}
        title="VapusAI Documentation"
        src={process.env.NODE_ENV === 'development' 
          ? `http://localhost:3001${docsPath}` 
          : `/docs-static${docsPath}`}
      />
    </div>
  );
}
