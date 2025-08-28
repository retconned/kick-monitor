import type React from "react";
import { useState } from "react";
import { Button } from "../components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Textarea } from "../components/ui/textarea";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "../components/ui/select";
import { Checkbox } from "../components/ui/checkbox";
import { Badge } from "../components/ui/badge";
import {
    Shield,
    ArrowLeft,
    Users,
    MessageSquare,
    Plus,
    CheckCircle,
} from "lucide-react";
import { NavLink } from "react-router";

export default function ContactPage() {
    const [isLoading, setIsLoading] = useState(false);
    const [isSubmitted, setIsSubmitted] = useState(false);
    const [requestType, setRequestType] = useState("");

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        // Simulate API call
        await new Promise((resolve) => setTimeout(resolve, 2000));
        setIsLoading(false);
        setIsSubmitted(true);
    };

    if (isSubmitted) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center p-4">
                {/* Grid Background */}
                <div className="absolute inset-0 bg-grid-white/[0.02] bg-grid-16" />

                <div className="relative w-full max-w-md text-center">
                    <div className="flex items-center justify-center mb-6">
                        <div className="h-16 w-16 rounded-full bg-green-500/20 flex items-center justify-center">
                            <CheckCircle className="h-8 w-8 text-green-500" />
                        </div>
                    </div>
                    <h1 className="text-2xl font-bold mb-4">
                        Request Submitted!
                    </h1>
                    <p className="text-muted-foreground mb-6">
                        Thank you for your request. We'll review it and get back
                        to you within 24-48 hours.
                    </p>
                    <div className="space-y-3">
                        <NavLink to="/">
                            <Button className="w-full bg-primary hover:bg-primary/90">
                                Back to Home
                            </Button>
                        </NavLink>
                        <Button
                            variant="outline"
                            className="w-full"
                            onClick={() => setIsSubmitted(false)}
                        >
                            Submit Another Request
                        </Button>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background p-4">
            {/* Grid Background */}
            <div className="absolute inset-0 bg-grid-white/[0.02] bg-grid-16" />

            {/* Floating Elements */}
            <div className="absolute top-20 left-10 w-20 h-20 bg-primary/10 rounded-full blur-xl" />
            <div className="absolute bottom-20 right-10 w-32 h-32 bg-primary/5 rounded-full blur-2xl" />

            <div className="relative max-w-4xl mx-auto">
                {/* Back to Home */}
                <NavLink
                    to="/"
                    className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground transition-colors mb-6"
                >
                    <ArrowLeft className="h-4 w-4 mr-2" />
                    Back to Home
                </NavLink>

                {/* Header */}
                <div className="text-center mb-12">
                    <div className="flex items-center justify-center space-x-2 mb-6">
                        <div className="h-10 w-10 rounded-lg bg-primary/20 flex items-center justify-center">
                            <Shield className="h-6 w-6 text-primary" />
                        </div>
                        <span className="text-2xl font-bold">KickMonitor</span>
                    </div>
                    <h1 className="text-4xl font-bold mb-4">Get in Touch</h1>
                    <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
                        Request new streamers to be added to our monitoring
                        system or get in touch with our team.
                    </p>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                    {/* Contact Info */}
                    <div className="space-y-6">
                        <Card className="bg-muted/30 border-muted">
                            <CardHeader>
                                <CardTitle className="flex items-center space-x-2">
                                    <Users className="h-5 w-5 text-primary" />
                                    <span>Request Streamer Monitoring</span>
                                </CardTitle>
                                <CardDescription>
                                    Want us to track a specific streamer? Let us
                                    know and we'll add them to our system.
                                </CardDescription>
                            </CardHeader>
                        </Card>

                        <Card className="bg-muted/30 border-muted">
                            <CardHeader>
                                <CardTitle className="flex items-center space-x-2">
                                    <MessageSquare className="h-5 w-5 text-primary" />
                                    <span>General Inquiries</span>
                                </CardTitle>
                                <CardDescription>
                                    Have questions about our platform, pricing,
                                    or features? We're here to help.
                                </CardDescription>
                            </CardHeader>
                        </Card>

                        <div className="space-y-4">
                            <h3 className="font-semibold">
                                Currently Monitoring
                            </h3>
                            <div className="flex flex-wrap gap-2">
                                <Badge variant="outline">xQc</Badge>
                                <Badge variant="outline">Ninja</Badge>
                                <Badge variant="outline">Pokimane</Badge>
                                <Badge variant="outline">Shroud</Badge>
                                <Badge variant="outline">TimTheTatman</Badge>
                                <Badge variant="outline">DrLupo</Badge>
                                <Badge variant="outline">LIRIK</Badge>
                                <Badge variant="outline">Summit1g</Badge>
                                <Badge
                                    variant="outline"
                                    className="text-primary border-primary"
                                >
                                    <Plus className="h-3 w-3 mr-1" />
                                    50K+ more
                                </Badge>
                            </div>
                        </div>
                    </div>

                    {/* Contact Form */}
                    <div className="lg:col-span-2">
                        <Card className="bg-muted/30 border-muted">
                            <CardHeader>
                                <CardTitle>Send us a message</CardTitle>
                                <CardDescription>
                                    Fill out the form below and we'll get back
                                    to you as soon as possible.
                                </CardDescription>
                            </CardHeader>
                            <CardContent>
                                <form
                                    onSubmit={handleSubmit}
                                    className="space-y-6"
                                >
                                    {/* Request Type */}
                                    <div className="space-y-2">
                                        <Label htmlFor="requestType">
                                            Request Type
                                        </Label>
                                        <Select
                                            value={requestType}
                                            onValueChange={setRequestType}
                                            required
                                        >
                                            <SelectTrigger className="bg-background/50">
                                                <SelectValue placeholder="Select request type" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="streamer-request">
                                                    Add Streamer to Monitoring
                                                </SelectItem>
                                                <SelectItem value="general-inquiry">
                                                    General Inquiry
                                                </SelectItem>
                                                <SelectItem value="technical-support">
                                                    Technical Support
                                                </SelectItem>
                                                <SelectItem value="partnership">
                                                    Partnership Opportunity
                                                </SelectItem>
                                                <SelectItem value="feedback">
                                                    Feedback & Suggestions
                                                </SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>

                                    {/* Personal Info */}
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <div className="space-y-2">
                                            <Label htmlFor="firstName">
                                                First Name
                                            </Label>
                                            <Input
                                                id="firstName"
                                                type="text"
                                                placeholder="John"
                                                required
                                                className="bg-background/50"
                                            />
                                        </div>
                                        <div className="space-y-2">
                                            <Label htmlFor="lastName">
                                                Last Name
                                            </Label>
                                            <Input
                                                id="lastName"
                                                type="text"
                                                placeholder="Doe"
                                                required
                                                className="bg-background/50"
                                            />
                                        </div>
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="email">Email</Label>
                                        <Input
                                            id="email"
                                            type="email"
                                            placeholder="john@example.com"
                                            required
                                            className="bg-background/50"
                                        />
                                    </div>

                                    {/* Conditional Fields */}
                                    {requestType === "streamer-request" && (
                                        <div className="space-y-4 p-4 bg-primary/5 rounded-lg border border-primary/20">
                                            <h3 className="font-semibold text-primary">
                                                Streamer Information
                                            </h3>
                                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                <div className="space-y-2">
                                                    <Label htmlFor="streamerName">
                                                        Streamer Name/Handle
                                                    </Label>
                                                    <Input
                                                        id="streamerName"
                                                        type="text"
                                                        placeholder="e.g., xQc, Ninja, etc."
                                                        required
                                                        className="bg-background/50"
                                                    />
                                                </div>
                                                <div className="space-y-2">
                                                    <Label htmlFor="platform">
                                                        Platform
                                                    </Label>
                                                    <Select required>
                                                        <SelectTrigger className="bg-background/50">
                                                            <SelectValue placeholder="Select platform" />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            <SelectItem value="twitch">
                                                                Twitch
                                                            </SelectItem>
                                                            <SelectItem value="youtube">
                                                                YouTube
                                                            </SelectItem>
                                                            <SelectItem value="kick">
                                                                Kick
                                                            </SelectItem>
                                                            <SelectItem value="other">
                                                                Other
                                                            </SelectItem>
                                                        </SelectContent>
                                                    </Select>
                                                </div>
                                            </div>
                                            <div className="space-y-2">
                                                <Label htmlFor="streamerUrl">
                                                    Channel URL
                                                </Label>
                                                <Input
                                                    id="streamerUrl"
                                                    type="url"
                                                    placeholder="https://twitch.tv/username"
                                                    className="bg-background/50"
                                                />
                                            </div>
                                            <div className="space-y-2">
                                                <Label htmlFor="reason">
                                                    Why should we monitor this
                                                    streamer?
                                                </Label>
                                                <Textarea
                                                    id="reason"
                                                    placeholder="Tell us why this streamer would be valuable to monitor..."
                                                    className="bg-background/50 min-h-[80px]"
                                                />
                                            </div>
                                        </div>
                                    )}

                                    <div className="space-y-2">
                                        <Label htmlFor="subject">Subject</Label>
                                        <Input
                                            id="subject"
                                            type="text"
                                            placeholder="Brief description of your request"
                                            required
                                            className="bg-background/50"
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="message">Message</Label>
                                        <Textarea
                                            id="message"
                                            placeholder="Provide more details about your request..."
                                            required
                                            className="bg-background/50 min-h-[120px]"
                                        />
                                    </div>

                                    {/* Priority */}
                                    <div className="space-y-2">
                                        <Label htmlFor="priority">
                                            Priority
                                        </Label>
                                        <Select>
                                            <SelectTrigger className="bg-background/50">
                                                <SelectValue placeholder="Select priority level" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="low">
                                                    Low - General inquiry
                                                </SelectItem>
                                                <SelectItem value="medium">
                                                    Medium - Standard request
                                                </SelectItem>
                                                <SelectItem value="high">
                                                    High - Urgent matter
                                                </SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>

                                    {/* Newsletter Signup */}
                                    <div className="flex items-center space-x-2">
                                        <Checkbox id="newsletter" />
                                        <Label
                                            htmlFor="newsletter"
                                            className="text-sm"
                                        >
                                            Subscribe to our newsletter for
                                            updates and new features
                                        </Label>
                                    </div>

                                    <Button
                                        type="submit"
                                        className="w-full bg-primary hover:bg-primary/90"
                                        disabled={isLoading}
                                    >
                                        {isLoading
                                            ? "Sending..."
                                            : "Send Message"}
                                    </Button>
                                </form>
                            </CardContent>
                        </Card>
                    </div>
                </div>

                {/* FAQ Section */}
                <div className="mt-16">
                    <h2 className="text-2xl font-bold text-center mb-8">
                        Frequently Asked Questions
                    </h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <Card className="bg-muted/30 border-muted">
                            <CardHeader>
                                <CardTitle className="text-lg">
                                    How long does it take to add a streamer?
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <p className="text-muted-foreground">
                                    We typically add new streamers within 24-48
                                    hours of receiving a request. Popular
                                    streamers are prioritized.
                                </p>
                            </CardContent>
                        </Card>

                        <Card className="bg-muted/30 border-muted">
                            <CardHeader>
                                <CardTitle className="text-lg">
                                    What platforms do you support?
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <p className="text-muted-foreground">
                                    We currently support Twitch, YouTube Live,
                                    and Kick. We're working on adding more
                                    platforms based on user demand.
                                </p>
                            </CardContent>
                        </Card>

                        <Card className="bg-muted/30 border-muted">
                            <CardHeader>
                                <CardTitle className="text-lg">
                                    Is there a limit to streamer requests?
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <p className="text-muted-foreground">
                                    Free users can request up to 5 streamers per
                                    month. Premium users have unlimited
                                    requests.
                                </p>
                            </CardContent>
                        </Card>

                        <Card className="bg-muted/30 border-muted">
                            <CardHeader>
                                <CardTitle className="text-lg">
                                    Do you monitor small streamers?
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <p className="text-muted-foreground">
                                    Yes! We monitor streamers of all sizes.
                                    However, we prioritize based on community
                                    interest and engagement.
                                </p>
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        </div>
    );
}
