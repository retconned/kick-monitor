// src/lib/validation.ts

import * as z from "zod";

export const loginSchema = z.object({
    email: z.string().email({ message: "Please enter a valid email address." }),
    password: z
        .string()
        .min(8, { message: "Password must be at least 8 characters long." }),
});

export const signupSchema = z
    .object({
        email: z.string().email({ message: "Please enter a valid email." }),
        password: z
            .string()
            .min(8, { message: "Password must be at least 8 characters." }),
        confirmPassword: z.string(),
    })
    .refine((data) => data.password === data.confirmPassword, {
        message: "Passwords do not match.",
        path: ["confirmPassword"],
    });

// Type inference for easy use in our components
export type TLoginSchema = z.infer<typeof loginSchema>;
export type TSignupSchema = z.infer<typeof signupSchema>;
