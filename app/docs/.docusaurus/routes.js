import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  {
    path: '/docs-static/',
    component: ComponentCreator('/docs-static/', 'bdb'),
    exact: true
  },
  {
    path: '/docs-static/',
    component: ComponentCreator('/docs-static/', '7f9'),
    routes: [
      {
        path: '/docs-static/',
        component: ComponentCreator('/docs-static/', 'c58'),
        routes: [
          {
            path: '/docs-static/',
            component: ComponentCreator('/docs-static/', 'cda'),
            routes: [
              {
                path: '/docs-static/features/model-registry',
                component: ComponentCreator('/docs-static/features/model-registry', '42d'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/docs-static/getting-started/installation/local',
                component: ComponentCreator('/docs-static/getting-started/installation/local', '960'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/docs-static/getting-started/installation/production',
                component: ComponentCreator('/docs-static/getting-started/installation/production', 'df5'),
                exact: true,
                sidebar: "tutorialSidebar"
              },
              {
                path: '/docs-static/getting-started/overview',
                component: ComponentCreator('/docs-static/getting-started/overview', '8a1'),
                exact: true,
                sidebar: "tutorialSidebar"
              }
            ]
          }
        ]
      }
    ]
  },
  {
    path: '*',
    component: ComponentCreator('*'),
  },
];
