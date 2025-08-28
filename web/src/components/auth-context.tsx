import React, { createContext, useState, useEffect } from "react";
import { useNavigate } from "react-router";

interface AuthContextType {
    isAuthenticated: boolean;
    isLoading: boolean;
    login: (token: string) => void;
    logout: () => void;
}

export const AuthContext = createContext<AuthContextType | undefined>(
    undefined,
);

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [isLoading, setIsLoading] = useState(true);
    const navigate = useNavigate();

    useEffect(() => {
        // On app load, check for a token in localStorage
        const token = localStorage.getItem("token");
        if (token) {
            setIsAuthenticated(true);
        }
        setIsLoading(false);
    }, []);

    const login = (token: string) => {
        localStorage.setItem("token", token);
        setIsAuthenticated(true);
        navigate("/admin");
    };

    const logout = () => {
        // Logout is now a purely client-side action
        localStorage.removeItem("token");
        setIsAuthenticated(false);
        navigate("/auth");
    };

    return (
        <AuthContext.Provider
            value={{ isAuthenticated, isLoading, login, logout }}
        >
            {children}
        </AuthContext.Provider>
    );
}
