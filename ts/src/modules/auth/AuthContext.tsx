import { createContext } from 'react';

interface AuthContextType { }

export const AuthContext = createContext<AuthContextType | null>(null);

export default AuthContext;
