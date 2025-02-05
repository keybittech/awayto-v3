import { createContext } from 'react';
import Keycloak from 'keycloak-js';

declare global {
  type AuthContextType = {
    authenticated: boolean;
    keycloak: Keycloak;
    token?: string;
  }
}

export const AuthContext = createContext<AuthContextType | null>(null);

export default AuthContext;
