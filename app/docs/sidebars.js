/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  tutorialSidebar: [
    {
      type: 'category',
      label: 'Getting Started',
      collapsed: false,
      items: [
        'getting-started/overview',
        {
          type: 'category',
          label: 'Installation',
          collapsed: false,
          items: [
            'getting-started/installation/local',
            'getting-started/installation/production',
          ],
        },
      ],
    },
    {
      type: 'category',
      label: 'Features',
      collapsed: false,
      items: ['features/model-registry'],
    },
  ],
};

export default sidebars;
