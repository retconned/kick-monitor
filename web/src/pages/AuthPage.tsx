import { useState } from "react";
import { NavLink } from "react-router";
import { Button } from "../components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "../components/ui/card";
import { Input } from "../components/ui/input";
import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
} from "../components/ui/tabs";

import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "../components/ui/form";

import { Shield, Eye, EyeOff, ArrowLeft, Loader2 } from "lucide-react";

import { useAuth } from "../hooks/useAuth";
import { Toaster, toast } from "sonner";
import { useMutation } from "@tanstack/react-query";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import {
    loginSchema,
    signupSchema,
    type TLoginSchema,
    type TSignupSchema,
} from "../lib/validation";
import { apiFetch } from "../lib/api";

interface LoginResponse {
    message: string;
    token: string;
}

const loginUser = (data: TLoginSchema) => {
    // This function now expects a LoginResponse
    return apiFetch<LoginResponse>("/login", {
        method: "POST",
        body: JSON.stringify(data),
    });
};

const registerUser = async (data: TSignupSchema) => {
    const payload = { email: data.email, password: data.password };

    const response = await fetch("http://localhost:8080/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
    });

    if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || "Failed to register.");
    }
    return response.json();
};

function LoginForm() {
    const { login } = useAuth();
    const [showPassword, setShowPassword] = useState(false);

    const form = useForm<TLoginSchema>({
        resolver: zodResolver(loginSchema),
        defaultValues: { email: "", password: "" },
    });

    const { mutate: performLogin, isPending } = useMutation({
        mutationFn: loginUser,
        onSuccess: (data) => {
            // Extract the token from the response and pass it to the context
            toast.success("Login successful! Redirecting...");
            login(data.token);
        },
        onError: (error: Error) => {
            toast.error(error.message);
        },
    });

    return (
        <Form {...form}>
            <form
                onSubmit={form.handleSubmit((data) => performLogin(data))}
                className="space-y-4"
            >
                <FormField
                    control={form.control}
                    name="email"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Email</FormLabel>
                            <FormControl>
                                <Input
                                    placeholder="Enter your email"
                                    {...field}
                                    className="bg-background/50"
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <FormField
                    control={form.control}
                    name="password"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Password</FormLabel>
                            <FormControl>
                                <div className="relative">
                                    <Input
                                        type={
                                            showPassword ? "text" : "password"
                                        }
                                        placeholder="Enter your password"
                                        {...field}
                                        className="bg-background/50 pr-10"
                                    />
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="sm"
                                        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                                        onClick={() =>
                                            setShowPassword(!showPassword)
                                        }
                                    >
                                        {showPassword ? (
                                            <EyeOff className="h-4 w-4 text-muted-foreground" />
                                        ) : (
                                            <Eye className="h-4 w-4 text-muted-foreground" />
                                        )}
                                    </Button>
                                </div>
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <Button type="submit" className="w-full" disabled={isPending}>
                    {isPending && (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    )}
                    {isPending ? "Signing in..." : "Sign In"}
                </Button>
            </form>
        </Form>
    );
}

function SignupForm() {
    const { login } = useAuth();
    const [showPassword, setShowPassword] = useState(false);

    const form = useForm<TSignupSchema>({
        resolver: zodResolver(signupSchema),
        defaultValues: { email: "", password: "", confirmPassword: "" },
    });

    const { mutate: performRegister, isPending } = useMutation({
        mutationFn: registerUser,
        onSuccess: () => {
            toast.success("Account created! Logging you in...");
            // After successful registration, automatically log the user in
            const loginData = form.getValues();
            performLoginAfterRegister(loginData);
        },
        onError: (error) => {
            toast.error(error.message);
        },
    });

    const { mutate: performLoginAfterRegister } = useMutation({
        mutationFn: loginUser,
        onSuccess: (data) => {
            login(data.token);
        },
        onError: (error: Error) => {
            toast.error(`Login after signup failed: ${error.message}`);
        },
    });

    return (
        <Form {...form}>
            <form
                onSubmit={form.handleSubmit((data) => performRegister(data))}
                className="space-y-4"
            >
                {/* Add First/Last Name fields here if needed */}
                <FormField
                    control={form.control}
                    name="email"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Email</FormLabel>
                            <FormControl>
                                <Input
                                    placeholder="Enter your email"
                                    {...field}
                                    className="bg-background/50"
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <FormField
                    control={form.control}
                    name="password"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Password</FormLabel>
                            <FormControl>
                                <div className="relative">
                                    <Input
                                        type={
                                            showPassword ? "text" : "password"
                                        }
                                        placeholder="Create a password"
                                        {...field}
                                        className="bg-background/50 pr-10"
                                    />
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="sm"
                                        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                                        onClick={() =>
                                            setShowPassword(!showPassword)
                                        }
                                    >
                                        {showPassword ? (
                                            <EyeOff className="h-4 w-4 text-muted-foreground" />
                                        ) : (
                                            <Eye className="h-4 w-4 text-muted-foreground" />
                                        )}
                                    </Button>
                                </div>
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <FormField
                    control={form.control}
                    name="confirmPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Confirm Password</FormLabel>
                            <FormControl>
                                <Input
                                    type={showPassword ? "text" : "password"}
                                    placeholder="Confirm your password"
                                    {...field}
                                    className="bg-background/50"
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <Button type="submit" className="w-full" disabled={isPending}>
                    {isPending && (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    )}
                    {isPending ? "Creating account..." : "Create Account"}
                </Button>
            </form>
        </Form>
    );
}

export default function AuthPage() {
    return (
        <>
            <Toaster richColors position="top-center" />
            <div className="min-h-screen bg-background flex items-center justify-center p-4">
                <div className="absolute inset-0 bg-grid-white/[0.02] bg-grid-16" />
                <div className="relative w-full max-w-md">
                    <NavLink
                        to="/"
                        className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground transition-colors mb-6"
                    >
                        <ArrowLeft className="h-4 w-4 mr-2" />
                        Back to Home
                    </NavLink>

                    <div className="flex items-center justify-center space-x-2 mb-8">
                        <div className="h-10 w-10 rounded-lg bg-primary/20 flex items-center justify-center">
                            <Shield className="h-6 w-6 text-primary" />
                        </div>
                        <span className="text-2xl font-bold">KickMonitor</span>
                    </div>

                    <Card className="bg-muted/30 border-muted backdrop-blur-sm">
                        <CardHeader className="text-center pb-4">
                            <CardTitle className="text-2xl">Welcome</CardTitle>
                            <CardDescription>
                                Sign in to your account or create a new one
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Tabs defaultValue="login" className="space-y-6">
                                <TabsList className="grid w-full grid-cols-2 bg-muted/50">
                                    <TabsTrigger value="login">
                                        Sign In
                                    </TabsTrigger>
                                    <TabsTrigger value="signup">
                                        Sign Up
                                    </TabsTrigger>
                                </TabsList>
                                <TabsContent value="login">
                                    <LoginForm />
                                </TabsContent>
                                <TabsContent value="signup">
                                    <SignupForm />
                                </TabsContent>
                            </Tabs>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </>
    );
}
