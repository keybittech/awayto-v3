import { createContext } from 'react';

declare global {
  type AuthContextType = {}
}

export const AuthContext = createContext<AuthContextType | null>(null);

export default AuthContext;
