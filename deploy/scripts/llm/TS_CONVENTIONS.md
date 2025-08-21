# Awayto TypeScript Project Guide

This guide outlines the architecture, API construction, module structure, and coding conventions for the TypeScript frontend application located in the `ts/` directory.

## 1. Project Overview & Architecture

The `ts/` directory houses the React frontend application, built upon ReduxJS and leveraging a Protobuf-first API design. This architecture ensures strong typing and consistency between the frontend and the Go backend.

## 2. API Construction with Protobufs & RTK-Query

The API for the TypeScript frontend is automatically generated, providing a streamlined and type-safe interaction with the backend.

- **Protobufs as Foundation**: All HTTP services and data types are defined in `.proto` files located in the top-level `protos/` folder. These files serve as the single source of truth for API contracts.

- **OpenAPI Specification Generation**: A `protoc` command (e.g., `protoc --proto_path=proto --experimental_allow_proto3_optional --openapi_out=$(TS_SRC) $(PROTO_FILES)`) is used to convert these Protobuf service definitions into `ts/openapi.yaml`. This `openapi.yaml` file describes the API endpoints.

- **RTK-Query Hook Generation**: Subsequently, `npx -y @rtk-query/codegen-openapi $(TS_CONFIG_API)` processes the `openapi.yaml` to auto-generate RTK-Query hooks in `ts/hooks/api.ts`.

- **`siteApi` Variable**: The generated API is exposed through the `siteApi` variable, which can be imported from the `'awayto-hooks'` package. This `siteApi` contains all the necessary RTK-Query hooks for interacting with the backend.
  **IMPORTANT**: `siteApi` and its underlying `api.ts` file are auto-generated and should NEVER be manually modified. Developers should only utilize its methods.

  **Example Usage of `siteApi` in a Component**:
  Here's an example of how to fetch a list of users and display them in a React component using the `siteApi`'s generated hooks.

  ```typescript
  import React from 'react';

  import CircularProgress from '@mui/material/CircularProgress';
  import List from '@mui/material/List';
  import ListItem from '@mui/material/ListItem';
  import ListItemText from '@mui/material/ListItemText';
  import Typography from '@mui/material/Typography';

  import { siteApi } from 'awayto-hooks';

  interface UserListComponentProps extends IComponent {}

  export function UserListComponent({}: UserListComponentProps): React.JSX.Element {
    const { data: users, isLoading, isError, error } = siteApi.useGetUsersQuery({});

    if (isLoading) {
      return <CircularProgress />;
    }

    if (isError) {
      return <Typography color="error">Error loading users: {error?.message || 'Unknown error'}</Typography>;
    }

    return (
      <div>
        <Typography variant="h6">Users List</Typography>
        {users && users.length > 0 ? (
          <List>
            {users.map((user) => (
              <ListItem key={user.id}>
                <ListItemText primary={user.username} secondary={user.email} />
              </ListItem>
            ))}
          </List>
        ) : (
          <Typography>No users found.</Typography>
        )}
      </div>
    );
  }
  ```

## 3. Module and Component Structure

The `ts/src/modules/` directory contains distinct application modules, each encapsulating related components and logic.

- **Module Construction**: Modules typically group components that relate to a specific domain (e.g., `bookings`, `exchange`, `forms`, `groups`).

- **Component Structure**: React components within these modules often extend a base `IComponent` interface (defined in `ts/src/index.tsx`), which provides common props like `children`, `loading`, and `closeModal`.

  Example Component Signature (from `ts/src/modules/common/SelectLookup.tsx`):
  ```typescript
  interface SelectLookupProps extends IComponent {
    multiple?: boolean;
    required?: boolean;
    disabled?: boolean;
    noEmptyValue?: boolean;
    lookups?: (ILookup | undefined)[];
    lookupName: string;
    helperText: React.JSX.Element | string;
    lookupChange(value: string | string[]): void;
    lookupValue: string | string[];
  }
  ```

  Example Functional Component (from `ts/src/modules/ext/Ext.tsx`):
  ```typescript
  export function Ext(): React.JSX.Element {
    return <>
      <Routes>
        <Route path="/kiosk/:groupName/:scheduleName?" element={
          <Kiosk />
        } />
      </Routes>
    </>
  }
  ```

## 4. Coding Conventions

Adhering to these conventions ensures consistency and maintainability across the `ts/` codebase:

- **Hooks and Utilities**: All core application hooks and utilities should be imported from the `'awayto-hooks'` package. This package consolidates functionality from `ts/src/hooks/`.

- **Data Fetching**: All data fetching operations must exclusively use the `siteApi` variable and its generated RTK-Query hooks.

- **Material-UI Imports**: When importing Material-UI components, always use direct, third-level imports to avoid larger bundle sizes and maintain clarity.
  **NEVER** use destructuring/braces for MUI component imports.
  Correct: `import Typography from '@mui/material/Typography';`
  Incorrect: `import { Typography } from '@mui/material';`

- **Standard React Hooks**: Utilize standard React hooks (`useState`, `useEffect`, `useContext`, `useMemo`, `useCallback`, `useRef`, etc.) for managing component state and side effects.

- **Import Ordering**: Maintain a consistent import order:
  1.  Core libraries (e.g., `react`, `redux`).
  2.  Material-UI components (separated by a newline).
  3.  Awayto framework imports (e.g., `'awayto-hooks'`, separated by a newline).

- **TypeScript Usage**: Leverage TypeScript's type safety by defining interfaces for component props (e.g., `interface MyComponentProps extends IComponent { ... }`), function parameters, and return types, ensuring robust and predictable code.
