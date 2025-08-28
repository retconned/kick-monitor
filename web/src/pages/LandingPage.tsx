import { Button } from "../components/ui/button";
import { Card, CardContent } from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import {
    ArrowRight,
    Shield,
    Eye,
    AlertTriangle,
    Users,
    Zap,
    Target,
} from "lucide-react";

import { NavLink } from "react-router";
const features = [
    {
        icon: <Shield className="h-6 w-6" />,
        title: "Bot Detection Engine",
        description:
            "Advanced algorithms identify fake viewers, followers, and engagement patterns in real-time.",
    },
    {
        icon: <AlertTriangle className="h-6 w-6" />,
        title: "Fraud Alerts",
        description:
            "Get instant notifications when suspicious activity or artificial inflation is detected.",
    },
    {
        icon: <Eye className="h-6 w-6" />,
        title: "Authenticity Score",
        description:
            "See a real-time authenticity rating for any streamer based on genuine engagement metrics.",
    },
    {
        icon: <Target className="h-6 w-6" />,
        title: "Viewbot Analysis",
        description:
            "Detailed breakdowns of organic vs artificial viewers with timestamp accuracy.",
    },
    {
        icon: <Users className="h-6 w-6" />,
        title: "Follower Verification",
        description:
            "Distinguish between real followers and purchased bot accounts across all platforms.",
    },
    {
        icon: <Zap className="h-6 w-6" />,
        title: "Live Monitoring",
        description:
            "24/7 surveillance of streaming metrics to catch inflation attempts as they happen.",
    },
];

const stats = [
    { value: "2.3M+", label: "Bots Detected" },
    { value: "15K+", label: "Streamers Analyzed" },
    { value: "98.7%", label: "Detection Accuracy" },
    { value: "Real-time", label: "Fraud Detection" },
];

const testimonials = [
    {
        name: "AdAgency Pro",
        handle: "@adagencypro",
        content:
            "This saved us $50K in wasted ad spend. We can now verify authentic audiences before sponsoring streamers.",
        avatar: "/placeholder.svg?height=40&width=40",
    },
    {
        name: "TwitchPartner",
        handle: "@realstreamer",
        content:
            "Finally, a way to prove my audience is 100% real. This helps me stand out from streamers using viewbots.",
        avatar: "/placeholder.svg?height=40&width=40",
    },
    {
        name: "EsportsOrg",
        handle: "@esportsorg",
        content:
            "Essential tool for vetting potential team members. We've caught multiple fake influencers trying to join.",
        avatar: "/placeholder.svg?height=40&width=40",
    },
];

const detectionMethods = [
    {
        title: "Behavioral Analysis",
        description:
            "AI patterns detect unnatural viewer behavior and engagement timing",
        percentage: "94%",
    },
    {
        title: "Network Fingerprinting",
        description:
            "Identify bot farms and proxy networks used for view inflation",
        percentage: "97%",
    },
    {
        title: "Chat Analysis",
        description: "Distinguish between real users and automated chat bots",
        percentage: "91%",
    },
    {
        title: "Follower Patterns",
        description:
            "Detect mass follow/unfollow campaigns and purchased followers",
        percentage: "99%",
    },
];

export default function LandingPage() {
    return (
        <div className="min-h-screen bg-background">
            {/* Grid Background */}
            <div className="absolute inset-0 bg-grid-white/[0.02] bg-grid-16" />

            {/* Header */}
            <header className="relative border-b border-border/40 bg-background/80 backdrop-blur-sm">
                <div className="container mx-auto px-6 py-4">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                            <div className="h-8 w-8 rounded-lg bg-primary/20 flex items-center justify-center">
                                <Shield className="h-5 w-5 text-primary" />
                            </div>
                            <span className="text-xl font-bold">
                                KickMonitor
                            </span>
                        </div>
                        <nav className="hidden md:flex items-center space-x-8">
                            <a
                                href="#detection"
                                className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                            >
                                Detection
                            </a>
                            <a
                                href="#pricing"
                                className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                            >
                                Pricing
                            </a>
                            <NavLink
                                to="/contact"
                                className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                            >
                                Contact
                            </NavLink>
                        </nav>
                        <div className="flex items-center space-x-4">
                            <NavLink to="/auth">
                                <Button
                                    size="sm"
                                    className="bg-primary hover:bg-primary/90"
                                >
                                    Login
                                    <ArrowRight className="ml-2 h-4 w-4" />
                                </Button>
                            </NavLink>
                        </div>
                    </div>
                </div>
            </header>

            {/* Hero Section */}
            <section className="relative py-24 lg:py-32">
                <div className="container mx-auto px-6">
                    <div className="text-center max-w-4xl mx-auto">
                        <Badge
                            variant="outline"
                            className="mb-6 bg-red-500/10 text-red-500 border-red-500/20"
                        >
                            <AlertTriangle className="h-3 w-3 mr-1" />
                            2.3M+ Viewbots Detected
                        </Badge>

                        <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold tracking-tight mb-6">
                            We catch{" "}
                            <span className="bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
                                fake viewers and bots
                            </span>{" "}
                            so you know what's real.
                        </h1>

                        <p className="text-xl text-muted-foreground mb-8 max-w-2xl mx-auto leading-relaxed">
                            Advanced bot detection technology that exposes
                            artificial view inflation, fake followers, and
                            fraudulent engagement. Protect your investments and
                            discover authentic streamers with 98.7% accuracy.
                        </p>

                        <div className="flex flex-col sm:flex-row gap-4 justify-center">
                            <NavLink to="/admin">
                                <Button
                                    size="lg"
                                    className="bg-primary hover:bg-primary/90 text-primary-foreground px-8"
                                >
                                    Start Detection
                                    <Shield className="ml-2 h-5 w-5" />
                                </Button>
                            </NavLink>
                            <Button
                                variant="outline"
                                size="lg"
                                className="px-8"
                            >
                                See Live Demo
                            </Button>
                        </div>
                    </div>
                </div>

                {/* Floating Elements */}
                <div className="absolute top-20 left-10 w-20 h-20 bg-red-500/10 rounded-full blur-xl" />
                <div className="absolute bottom-20 right-10 w-32 h-32 bg-primary/5 rounded-full blur-2xl" />
            </section>

            {/* Stats Section */}
            <section className="py-16 border-y border-border/40 bg-muted/20">
                <div className="container mx-auto px-6">
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
                        {stats.map((stat, index) => (
                            <div key={index} className="text-center">
                                <div className="text-3xl md:text-4xl font-bold text-primary mb-2">
                                    {stat.value}
                                </div>
                                <div className="text-sm text-muted-foreground">
                                    {stat.label}
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </section>

            {/* Detection Methods */}
            <section id="detection" className="py-24">
                <div className="container mx-auto px-6">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl md:text-4xl font-bold mb-4">
                            <span className="text-primary">
                                Advanced detection
                            </span>{" "}
                            methods
                        </h2>
                        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
                            Our AI-powered system uses multiple detection
                            vectors to identify fraudulent activity with
                            industry-leading accuracy.
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-16">
                        {detectionMethods.map((method, index) => (
                            <Card
                                key={index}
                                className="bg-muted/30 border-muted hover:bg-muted/40 transition-colors group"
                            >
                                <CardContent className="p-6">
                                    <div className="flex items-center justify-between mb-4">
                                        <h3 className="text-lg font-semibold">
                                            {method.title}
                                        </h3>
                                        <Badge
                                            variant="outline"
                                            className="bg-primary/10 text-primary border-primary/20"
                                        >
                                            {method.percentage} accurate
                                        </Badge>
                                    </div>
                                    <p className="text-muted-foreground leading-relaxed">
                                        {method.description}
                                    </p>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                </div>
            </section>

            {/* Features Section */}
            <section className="py-24 bg-muted/20">
                <div className="container mx-auto px-6">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl md:text-4xl font-bold mb-4">
                            Everything you need to{" "}
                            <span className="text-primary">detect fraud</span>
                        </h2>
                        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
                            Comprehensive fraud detection tools designed for
                            advertisers, sponsors, and authentic content
                            creators.
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
                        {features.map((feature, index) => (
                            <Card
                                key={index}
                                className="bg-muted/30 border-muted hover:bg-muted/40 transition-colors group"
                            >
                                <CardContent className="p-6">
                                    <div className="flex items-center space-x-4 mb-4">
                                        <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary group-hover:bg-primary/20 transition-colors">
                                            {feature.icon}
                                        </div>
                                        <h3 className="text-lg font-semibold">
                                            {feature.title}
                                        </h3>
                                    </div>
                                    <p className="text-muted-foreground leading-relaxed">
                                        {feature.description}
                                    </p>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                </div>
            </section>

            {/* Dashboard Preview */}
            <section className="py-24">
                <div className="container mx-auto px-6">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl md:text-4xl font-bold mb-4">
                            Real-time{" "}
                            <span className="text-primary">
                                fraud detection
                            </span>{" "}
                            dashboard
                        </h2>
                        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
                            Monitor authenticity scores, detect bot activity,
                            and verify genuine engagement in real-time.
                        </p>
                    </div>

                    <div className="relative max-w-6xl mx-auto">
                        <div className="absolute inset-0 bg-gradient-to-r from-red-500/20 to-primary/20 rounded-2xl blur-3xl" />
                        <Card className="relative bg-muted/40 border-muted overflow-hidden">
                            <CardContent className="p-8">
                                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                                    <div className="lg:col-span-2">
                                        <div className="h-48 bg-gradient-to-br from-primary/20 to-red-500/10 rounded-lg flex items-center justify-center">
                                            <div className="text-center">
                                                <Shield className="h-12 w-12 text-primary mx-auto mb-4" />
                                                <p className="text-muted-foreground">
                                                    Live Bot Detection Analytics
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                    <div className="space-y-4">
                                        <div className="h-20 bg-muted/60 rounded-lg flex items-center justify-center">
                                            <div className="text-center">
                                                <div className="text-2xl font-bold text-green-500">
                                                    87%
                                                </div>
                                                <div className="text-xs text-muted-foreground">
                                                    Authentic Score
                                                </div>
                                            </div>
                                        </div>
                                        <div className="h-20 bg-muted/60 rounded-lg flex items-center justify-center">
                                            <div className="text-center">
                                                <div className="text-2xl font-bold text-red-500">
                                                    1,247
                                                </div>
                                                <div className="text-xs text-muted-foreground">
                                                    Bots Detected
                                                </div>
                                            </div>
                                        </div>
                                        <div className="h-20 bg-muted/60 rounded-lg flex items-center justify-center">
                                            <div className="text-center">
                                                <div className="text-2xl font-bold text-primary">
                                                    Real-time
                                                </div>
                                                <div className="text-xs text-muted-foreground">
                                                    Monitoring
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </section>

            {/* Use Cases */}
            <section className="py-24 bg-muted/20">
                <div className="container mx-auto px-6">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl md:text-4xl font-bold mb-4">
                            Who uses{" "}
                            <span className="text-primary">KickMonitor ?</span>
                        </h2>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                        <Card className="bg-muted/30 border-muted">
                            <CardContent className="p-6 text-center">
                                <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary mx-auto mb-4">
                                    <Target className="h-6 w-6" />
                                </div>
                                <h3 className="text-lg font-semibold mb-2">
                                    Advertisers & Sponsors
                                </h3>
                                <p className="text-muted-foreground">
                                    Verify authentic audiences before investing
                                    in influencer partnerships and sponsorship
                                    deals.
                                </p>
                            </CardContent>
                        </Card>

                        <Card className="bg-muted/30 border-muted">
                            <CardContent className="p-6 text-center">
                                <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary mx-auto mb-4">
                                    <Users className="h-6 w-6" />
                                </div>
                                <h3 className="text-lg font-semibold mb-2">
                                    Authentic Streamers
                                </h3>
                                <p className="text-muted-foreground">
                                    Prove your audience is real and stand out
                                    from competitors using artificial inflation.
                                </p>
                            </CardContent>
                        </Card>

                        <Card className="bg-muted/30 border-muted">
                            <CardContent className="p-6 text-center">
                                <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary mx-auto mb-4">
                                    <Target className="h-6 w-6" />
                                </div>
                                <h3 className="text-lg font-semibold mb-2">
                                    Platform Moderators
                                </h3>
                                <p className="text-muted-foreground">
                                    Identify and take action against accounts
                                    violating terms of service with fake
                                    engagement.
                                </p>
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </section>

            {/* Testimonials */}
            <section className="py-24">
                <div className="container mx-auto px-6">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl md:text-4xl font-bold mb-4">
                            Trusted by{" "}
                            <span className="text-primary">
                                industry leaders
                            </span>
                        </h2>
                        <p className="text-xl text-muted-foreground">
                            See how KickMonitor is protecting investments and
                            ensuring authenticity
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                        {testimonials.map((testimonial, index) => (
                            <Card
                                key={index}
                                className="bg-muted/30 border-muted"
                            >
                                <CardContent className="p-6">
                                    <p className="text-muted-foreground mb-4 leading-relaxed">
                                        "{testimonial.content}"
                                    </p>
                                    <div className="flex items-center space-x-3">
                                        <div className="h-10 w-10 rounded-full bg-primary/20 flex items-center justify-center">
                                            <span className="text-sm font-semibold text-primary">
                                                {testimonial.name.slice(0, 2)}
                                            </span>
                                        </div>
                                        <div>
                                            <div className="font-semibold text-sm">
                                                {testimonial.name}
                                            </div>
                                            <div className="text-xs text-muted-foreground">
                                                {testimonial.handle}
                                            </div>
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                </div>
            </section>

            {/* CTA Section */}
            <section className="py-24 bg-muted/20">
                <div className="container mx-auto px-6">
                    <div className="text-center max-w-3xl mx-auto">
                        <h2 className="text-3xl md:text-4xl font-bold mb-4">
                            Ready to{" "}
                            <span className="text-primary">
                                expose the fakes?
                            </span>
                        </h2>
                        <p className="text-xl text-muted-foreground mb-8">
                            Join thousands of advertisers and authentic creators
                            who trust KickMonitor to detect fraud.
                        </p>
                        <div className="flex flex-col sm:flex-row gap-4 justify-center">
                            <NavLink to="/admin">
                                <Button
                                    size="lg"
                                    className="bg-primary hover:bg-primary/90 text-primary-foreground px-8"
                                >
                                    Start Detection
                                    <Shield className="ml-2 h-5 w-5" />
                                </Button>
                            </NavLink>
                            <Button
                                variant="outline"
                                size="lg"
                                className="px-8"
                            >
                                Request Demo
                            </Button>
                        </div>
                    </div>
                </div>
            </section>

            {/* Footer */}
            <footer className="border-t border-border/40 py-12">
                <div className="container mx-auto px-6">
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
                        <div>
                            <div className="flex items-center space-x-2 mb-4">
                                <div className="h-8 w-8 rounded-lg bg-primary/20 flex items-center justify-center">
                                    <Shield className="h-5 w-5 text-primary" />
                                </div>
                                <span className="text-xl font-bold">
                                    KickMonitor
                                </span>
                            </div>
                            <p className="text-sm text-muted-foreground">
                                The most advanced bot detection platform for
                                streaming fraud prevention.
                            </p>
                        </div>
                        <div>
                            <h4 className="font-semibold mb-4">Detection</h4>
                            <ul className="space-y-2 text-sm text-muted-foreground">
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        Bot Detection
                                    </a>
                                </li>
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        Fraud Analysis
                                    </a>
                                </li>
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        API Access
                                    </a>
                                </li>
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        Real-time Alerts
                                    </a>
                                </li>
                            </ul>
                        </div>
                        <div>
                            <h4 className="font-semibold mb-4">Company</h4>
                            <ul className="space-y-2 text-sm text-muted-foreground">
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        About
                                    </a>
                                </li>
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        Research
                                    </a>
                                </li>
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        Careers
                                    </a>
                                </li>
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        Contact
                                    </a>
                                </li>
                            </ul>
                        </div>
                        <div>
                            <h4 className="font-semibold mb-4">Resources</h4>
                            <ul className="space-y-2 text-sm text-muted-foreground">
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        Documentation
                                    </a>
                                </li>
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        Case Studies
                                    </a>
                                </li>
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        Fraud Reports
                                    </a>
                                </li>
                                <li>
                                    <a
                                        href="#"
                                        className="hover:text-foreground transition-colors"
                                    >
                                        Privacy
                                    </a>
                                </li>
                            </ul>
                        </div>
                    </div>
                    <div className="border-t border-border/40 mt-8 pt-8 text-center text-sm text-muted-foreground">
                        <p>&copy; 2024 KickMonitor . All rights reserved.</p>
                    </div>
                </div>
            </footer>
        </div>
    );
}
