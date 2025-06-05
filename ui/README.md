# Olake Frontend UI with React + TypeScript + Ant Design Tailwind + Vite 🌈

## Design 🎨

Check out our design on Figma:
[Olake-App Figma Design](https://www.figma.com/design/FwLnU97I8LjtYNREPyYofc/Olake%2FDesign%2FCommunity?node-id=0-1&p=f&t=oOQMWCyXF6rzMzT1-0)

## Requirements ✅

- A latest LTS version of [Node.js](https://nodejs.org/en/download/).
- [pnpm](https://pnpm.io/installation), a fast, disk space efficient package manager for Node.js.

## Running the project locally 🚀

- 🌐 Clone the Repo using the SSH or HTTPS
- 🎉 Install all the required dependencies in the root directory of the project using pnpm. before that ensure that u have 'pnpm'.

```bash
pnpm install
```

- 🎮 Run the project from the root directory.

```bash
pnpm dev
```

### To format the code 🎭

```bash
pnpm format
```

## Checksum before commit && pushing to remote 🎭

- 📝 To check the `EsLint` issus

```bash
pnpm lint
```

- 🎨 To fix the `EsLint` issues

```bash
pnpm lint:fix
```

## Troubleshoot Options 🚨

- Run the following command to clean the `node_modules`. After that reinstall dependencies using `pnpm install`

```bash
pnpx npkill
```

## Folder Structure 📁

```text
├── public // Contains public resources. Ex favicon
│
├── src
│   ├── assets // Contains all assets used in App
│   ├── api  // AXIOS setup and mockdata
|       ├── services // API services for Jobs , Sources , Destinations
│   ├── modules  // Contains all modules of App
|   |   ├── auth // auth components
│   │   ├── common
|   |   |   ├── components // common components
|   |   |   ├── Modals // Modals used throughout the app
│   │   └── destinations // All destinations related  components
|   |   ├── sources // All sources related components
|   |   ├── jobs  // All jobs related components
|	|
│   ├── routes // routes
|   |
│   ├── store  // Contains all state management files
|   |   ├── authStore // auth specific states
|   |   ├── destinationStore // Destination specific states
|   |   ├── jobStore // Job specific states
|   |   ├── modalStore // Modals specific states
|   |   ├── selectionStore // state file for managing the selected entities
|   |   ├── sourceStore // Source specific states
|   |   ├── taskStore // Task specific states
|   |   ├── index // states root file
│   │
│   ├── types
|   |   ├── apiTypes // api specific types
|   |   ├── authTypes // auth specific types
|   |   ├── commonTypes // common types used across different files
|   |   ├── entityTypes // entity(source/destination) specific types
|   |   ├── streamTypes // stream specific types
|   |   ├── jobTypes // Job specific types
|   |   ├── destinationTypes //Destination specific types
|   |   ├── sourceTypes // Source specific types
|   |   ├── formTypes // Form specific types
|   |   ├── index  //types root file
|   |
|   ├── utils // common functions and constants
│   │
│   ├── main.tsx // Root File of App which has all providers
│   │
│   └── App.tsx // App
│
├── index.html // index file of React App
│
├── package.json
│
├── eslint.config.js // Configuration of ESLint Plugin
│
├── tsconfig.json // Configuration of TypeScript
│
├── tailwind.config.js // Configuration of Tailwind CSS
│
├── tsconfig.app.json // Configuration of TypeScript for Browser(App) Environment
│
├── tsconfig.node.json // Configuration of TypeScript for Node Environment
│
└── vite.config.ts // Configuration of Vite
```

## Used Packages 📦

- **Tailwind CSS**: Tailwind CSS is an open-source CSS framework.
- **zustand**: A small, fast, and scalable bearbones state management solution.
- **Axios**: Axios is a promise based HTTP client for browser and node.js.
- **React Router DOM**: Used for Routing.
- **Ant Design**: Component Library.
- **Phosphor-icons**: Icons Library for our App.
- **Vite**: Vite is a local development server
- **eslint/js**: ESLint Plugin is used for Linting.
- **prettier**: Prettier Plugin is used for Prettier issues.

## UX Tips

### Suspense Wrapper

- use `suspense` from `react` for Loading Animation.
