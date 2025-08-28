import { useState, useMemo } from "react";
import type React from "react";
import { useQuery } from "@tanstack/react-query";

import {
    ArrowLeft,
    MessageSquare,
    AlertTriangle,
    Copy,
    Flag,
    Eye,
    Clock,
    TrendingUp,
    Bot,
    ChevronLeft,
    ChevronRight,
    Loader2,
    AlertCircle,
    Monitor,
} from "lucide-react";
import { Button } from "../components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "../components/ui/select";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "../components/ui/table";
import {
    ChartContainer,
    ChartTooltip,
    ChartTooltipContent,
    type ChartConfig,
} from "../components/ui/chart";
import {
    Area,
    AreaChart,
    XAxis,
    YAxis,
    CartesianGrid,
    ResponsiveContainer,
} from "recharts";

import type { Livestream } from "../types/types";
import { useNavigate, useParams } from "react-router";

// Types
interface HeatmapDataPoint {
    hour: number;
    minute: number;
    count: number;
}

// Generate mock heatmap data
const generateMockHeatmapData = (): HeatmapDataPoint[] => {
    const data: HeatmapDataPoint[] = [];
    for (let hour = 14; hour <= 16; hour++) {
        for (let minute = 0; minute < 60; minute += 2) {
            data.push({
                hour,
                minute,
                count: Math.floor(Math.random() * 3000) + 500,
            });
        }
    }
    return data;
};

const mockHeatmapData = generateMockHeatmapData();

// Chart Configuration
const chartConfig = {
    viewers: {
        label: "Viewers",
        color: "oklch(0.6776 0.1481 238.1)",
    },
    messages: {
        label: "Messages",
        color: "oklch(0.6776 0.1481 238.1)",
    },
    count: {
        label: "Count",
        color: "oklch(0.6776 0.1481 238.1)",
    },
} satisfies ChartConfig;

// API Functions
const fetchLivestreamData = async (id: string): Promise<Livestream> => {
    const response = await fetch(`http://localhost:80/api/livestream/${id}`);

    if (!response.ok) {
        throw new Error(
            `Failed to fetch livestream data: ${response.statusText}`,
        );
    }

    const data = await response.json();
    return Array.isArray(data) ? data[0] : data;
};

// Custom Hook
const useLivestreamData = (id: string | undefined) => {
    return useQuery({
        queryKey: ["livestream", id],
        queryFn: () => fetchLivestreamData(id!),
        enabled: !!id,
        staleTime: 5 * 60 * 1000,
        retry: 3,
        retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    });
};

// Utility Functions
const formatTime = (timeString: string) => {
    const date = new Date(timeString);
    return date.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
        hour12: false,
    });
};

const formatDuration = (totalMins: number) => {
    const mins = totalMins % 60;
    const totalHours = Math.floor(totalMins / 60);

    if (totalHours === 0) {
        return `${mins}mins`;
    }

    const hours = totalHours % 24;
    const days = Math.floor(totalHours / 24);

    if (days === 0) {
        return `${hours}h${mins}mins`;
    }

    const dd = String(days).padStart(2, "0");
    return `${dd}day${hours}hours${mins}mins`;
};

const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return num.toLocaleString();
};

const getHeatmapColor = (count: number, maxCount: number) => {
    const intensity = count / maxCount;
    if (intensity > 0.8) return "bg-primary border-primary/40";
    if (intensity > 0.6) return "bg-primary/80 border-primary/30";
    if (intensity > 0.4) return "bg-primary/60 border-primary/25";
    if (intensity > 0.2) return "bg-primary/40 border-primary/20";
    if (intensity > 0.1) return "bg-primary/20 border-primary/15";
    return "bg-muted/30 border-muted";
};

// Components
const MetricCard = ({
    title,
    value,
    suffix = "",
    color = "text-foreground",
    icon,
    compact = false,
}: {
    title: string;
    value: number | string;
    suffix?: string;
    color?: string;
    icon?: React.ReactNode;
    compact?: boolean;
}) => (
    <Card className="bg-muted/30 border-muted">
        <CardContent className={compact ? "p-4" : "p-5"}>
            <div className="flex items-center justify-between mb-2">
                <p
                    className={`${compact ? "text-xs" : "text-sm"} font-medium text-muted-foreground`}
                >
                    {title}
                </p>
                {icon}
            </div>
            <div className="flex items-end justify-between">
                <div>
                    <p
                        className={`${compact ? "text-lg" : "text-2xl"} font-bold text-white ${color}`}
                    >
                        {typeof value === "number"
                            ? formatNumber(value)
                            : value}
                        {suffix}
                    </p>
                </div>
            </div>
        </CardContent>
    </Card>
);

const Pagination = ({
    currentPage,
    totalPages,
    onPageChange,
}: {
    currentPage: number;
    totalPages: number;
    onPageChange: (page: number) => void;
}) => {
    if (totalPages <= 1) return null;

    return (
        <div className="flex items-center justify-between">
            <div className="text-sm text-muted-foreground">
                Page {currentPage} of {totalPages}
            </div>
            <div className="flex items-center space-x-2">
                <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onPageChange(currentPage - 1)}
                    disabled={currentPage === 1}
                    className="h-8 w-8 p-0"
                >
                    <ChevronLeft className="h-4 w-4" />
                </Button>
                <div className="flex items-center space-x-1">
                    {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                        let pageNum;
                        if (totalPages <= 5) {
                            pageNum = i + 1;
                        } else if (currentPage <= 3) {
                            pageNum = i + 1;
                        } else if (currentPage >= totalPages - 2) {
                            pageNum = totalPages - 4 + i;
                        } else {
                            pageNum = currentPage - 2 + i;
                        }

                        return (
                            <Button
                                key={pageNum}
                                variant={
                                    currentPage === pageNum
                                        ? "default"
                                        : "outline"
                                }
                                size="sm"
                                onClick={() => onPageChange(pageNum)}
                                className="h-8 w-8 p-0 text-xs"
                            >
                                {pageNum}
                            </Button>
                        );
                    })}
                </div>
                <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onPageChange(currentPage + 1)}
                    disabled={currentPage === totalPages}
                    className="h-8 w-8 p-0"
                >
                    <ChevronRight className="h-4 w-4" />
                </Button>
            </div>
        </div>
    );
};

const LoadingSpinner = ({ className }: { className?: string }) => (
    <div className={`flex items-center justify-center ${className}`}>
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
    </div>
);

const ErrorDisplay = ({
    error,
    onRetry,
}: {
    error: Error;
    onRetry: () => void;
}) => (
    <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4">
        <AlertCircle className="h-12 w-12 text-destructive" />
        <div className="text-center">
            <h3 className="font-semibold text-lg mb-2">
                Failed to load livestream data
            </h3>
            <p className="text-muted-foreground mb-4">{error.message}</p>
            <Button onClick={onRetry} variant="outline">
                Try Again
            </Button>
        </div>
    </div>
);

const generateHeatmapDataFromTimeline = (
    timeline: { time: Date; count: number }[],
) => {
    if (!timeline || timeline.length === 0) return [];

    // Map the timeline data to the heatmap format
    // Assuming timeline `time` is already Date objects

    return timeline.map((item) => ({
        hour: new Date(item.time).getHours(),
        minute: new Date(item.time).getMinutes(),
        count: item.count,
    }));
};

// Main Component
export default function StreamDetailsPage() {
    const [exactDuplicateSort, setExactDuplicateSort] = useState("count");
    const [similarMessageSort, setSimilarMessageSort] = useState("count");
    const [suspiciousPage, setSuspiciousPage] = useState(1);
    const [exactDuplicatePage, setExactDuplicatePage] = useState(1);
    const [similarMessagePage, setSimilarMessagePage] = useState(1);
    const itemsPerPage = 5;

    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();

    const { data, isLoading, isError, error, refetch } = useLivestreamData(id);

    // Memoized computations with proper null checks
    const suspiciousChatters = useMemo(() => {
        return data?.spam_report?.suspicious_chatters || [];
    }, [data?.spam_report?.suspicious_chatters]);

    const exactDuplicateBursts = useMemo(() => {
        return data?.spam_report?.exact_duplicate_bursts || [];
    }, [data?.spam_report?.exact_duplicate_bursts]);

    const similarMessageBursts = useMemo(() => {
        return data?.spam_report?.similar_message_bursts || [];
    }, [data?.spam_report?.similar_message_bursts]);

    // Paginated data
    const paginatedSuspiciousChatters = useMemo(() => {
        const startIndex = (suspiciousPage - 1) * itemsPerPage;
        const endIndex = startIndex + itemsPerPage;
        return suspiciousChatters.slice(startIndex, endIndex);
    }, [suspiciousChatters, suspiciousPage, itemsPerPage]);

    const paginatedExactDuplicates = useMemo(() => {
        const sorted = [...exactDuplicateBursts];
        switch (exactDuplicateSort) {
            case "count":
                sorted.sort((a, b) => b.count - a.count);
                break;
            // case "bursts":
            //     sorted.sort((a, b) => b.timestamps - a.timestamps);
            //     break;
            // case "users":
            //     sorted.sort((a, b) => b.userCount - a.userCount);
            //     break;
            // case "avgSize":
            //     sorted.sort((a, b) => b.avgBurstSize - a.avgBurstSize);
            //     break;
        }
        const startIndex = (exactDuplicatePage - 1) * itemsPerPage;
        const endIndex = startIndex + itemsPerPage;
        return sorted.slice(startIndex, endIndex);
    }, [
        exactDuplicateBursts,
        exactDuplicateSort,
        exactDuplicatePage,
        itemsPerPage,
    ]);

    const paginatedSimilarMessages = useMemo(() => {
        const sorted = [...similarMessageBursts];
        switch (similarMessageSort) {
            case "count":
                sorted.sort((a, b) => b.count - a.count);
                break;
            // case "bursts":
            //     sorted.sort((a, b) => b.burstCount - a.burstCount);
            //     break;
            // case "users":
            //     sorted.sort((a, b) => b.userCount - a.userCount);
            //     break;
            // case "similarity":
            //     sorted.sort((a, b) => b.similarity - a.similarity);
            //     break;
        }
        const startIndex = (similarMessagePage - 1) * itemsPerPage;
        const endIndex = startIndex + itemsPerPage;
        return sorted.slice(startIndex, endIndex);
    }, [
        similarMessageBursts,
        similarMessageSort,
        similarMessagePage,
        itemsPerPage,
    ]);

    // Calculate total pages
    const suspiciousTotalPages = Math.ceil(
        suspiciousChatters.length / itemsPerPage,
    );
    const exactDuplicateTotalPages = Math.ceil(
        exactDuplicateBursts.length / itemsPerPage,
    );
    const similarMessageTotalPages = Math.ceil(
        similarMessageBursts.length / itemsPerPage,
    );

    // --- Heatmap data generated from actual message_counts_timeline ---
    const heatmapData = useMemo(() => {
        return generateHeatmapDataFromTimeline(
            data?.message_counts_timeline || [],
        );
    }, [data?.message_counts_timeline]);

    // Calculate max count for heatmap coloring
    const maxHeatmapCount = useMemo(() => {
        if (heatmapData.length === 0) return 1; // Avoid division by zero
        return Math.max(...heatmapData.map((d) => d.count));
    }, [heatmapData]);

    const handleGoBack = () => {
        navigate(-1);
    };

    // Handle loading state
    if (isLoading) {
        return (
            <div className="min-h-screen bg-muted/20 flex items-center justify-center">
                <LoadingSpinner />
            </div>
        );
    }

    // Handle error state
    if (isError || !data) {
        return (
            <div className="min-h-screen bg-muted/20 flex items-center justify-center">
                <ErrorDisplay
                    error={error as Error}
                    onRetry={() => refetch()}
                />
            </div>
        );
    }

    const {
        title,
        duration_minutes,
        average_viewers,
        peak_viewers,
        engagement,
        hours_watched,
        total_messages,
        created_at: date,
        viewer_counts_timeline,
        message_counts_timeline,
    } = data;

    // Determine the range of hours for the heatmap
    const startHour = new Date(data.report_start_time).getHours();
    const endHour = new Date(data.report_end_time).getHours();
    const heatmapHours: number[] = [];
    for (let h = startHour; h <= endHour; h++) {
        heatmapHours.push(h);
    }

    // If the stream ends past midnight, handle wrap-around for display
    if (
        new Date(data.report_end_time).getDate() >
        new Date(data.report_start_time).getDate()
    ) {
        const lastDayHour = endHour;
        for (let h = startHour; h < 24; h++) {
            // from startHour to midnight
            if (!heatmapHours.includes(h)) heatmapHours.push(h);
        }
        for (let h = 0; h <= lastDayHour; h++) {
            // from midnight to endHour
            if (!heatmapHours.includes(h)) heatmapHours.push(h);
        }
        heatmapHours.sort((a, b) => a - b); // Ensure sorted after adding
    }
    // Handle cases where stream starts/ends within the same hour for display too
    if (heatmapHours.length === 0 && startHour === endHour) {
        heatmapHours.push(startHour);
    }
    // Filter heatmapData to only include points within the actual stream duration
    const filteredHeatmapData = heatmapData.filter((d) => {
        const dTime = new Date(); // Using current date just for time components
        dTime.setHours(d.hour, d.minute, 0, 0);

        // Check if the timestamp of the data point falls within the stream's start and end times
        // This is a simplified check and might need more robust date/time comparisons
        // if streams span across multiple days and we want to precisely filter.
        // For general "within stream duration", comparing hours is a good start.

        // This simple filter below only checks if data points fall within the start and end HOUR.
        // If a stream runs 14:30 - 15:30, it will show 14:00 and 15:00 hours.
        // The individual blocks themselves represent the 2-minute interval.
        return d.hour >= startHour && d.hour <= endHour;
    });

    console.log(filteredHeatmapData);
    console.log("heamaphours:", heatmapHours);

    // Calculate Avg Activity for heatmap
    // const avgHeatmapActivity = useMemo(() => {
    //     if (filteredHeatmapData.length === 0) return 0;
    //     const total = filteredHeatmapData.reduce((sum, d) => sum + d.count, 0);
    //     return Math.round(total / filteredHeatmapData.length);
    // }, [filteredHeatmapData]);

    return (
        <div className="min-h-screen bg-muted/20 flex flex-col lg:flex-row">
            {/* Side Panel */}
            <div className="w-full lg:w-96 bg-muted/40 border-b lg:border-b-0 lg:border-r border-muted p-5 overflow-y-auto">
                {/* Back Button */}
                <div className="mb-6">
                    <Button
                        onClick={handleGoBack}
                        variant="outline"
                        size="sm"
                        className="w-full bg-transparent"
                    >
                        <ArrowLeft className="h-4 w-4 mr-2" />
                        Go back
                    </Button>
                </div>

                {/* Stream Details */}
                <div className="mb-6">
                    <h2 className="text-lg font-semibold mb-4">
                        STREAM DETAILS
                    </h2>
                    <div className="space-y-3">
                        <div className="p-3 bg-muted/20 rounded-lg">
                            <h3 className="font-medium text-sm mb-2">
                                {title}
                            </h3>
                            <div className="space-y-1 text-xs text-muted-foreground">
                                <div className="flex justify-between">
                                    <span>Date:</span>
                                    <span>{new Date(date).toDateString()}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Stream Metrics */}
                <div>
                    <h2 className="text-lg font-semibold mb-4">METRICS</h2>
                    <div className="space-y-3">
                        <MetricCard
                            title="Avg viewers"
                            value={average_viewers}
                            color="text-primary"
                            icon={<Eye className="h-4 w-4 text-primary" />}
                            compact
                        />
                        <MetricCard
                            title="Peak viewers"
                            value={peak_viewers}
                            color="text-yellow-400"
                            icon={
                                <TrendingUp className="h-4 w-4 text-yellow-400" />
                            }
                            compact
                        />
                        <MetricCard
                            title="Total messages"
                            value={total_messages}
                            color="text-blue-400"
                            icon={
                                <MessageSquare className="h-4 w-4 text-blue-400" />
                            }
                            compact
                        />
                        <MetricCard
                            title="Duration"
                            value={formatDuration(duration_minutes)}
                            icon={<Clock className="h-4 w-4 text-pink-400" />}
                            compact
                        />
                        <MetricCard
                            title="Hours watched"
                            value={hours_watched}
                            icon={
                                <Monitor className="h-4 w-4 text-indigo-400" />
                            }
                            compact
                        />
                        <MetricCard
                            title="Engagement rate"
                            value={engagement}
                            icon={<Bot className="h-4 w-4 text-rose-400" />}
                            compact
                        />
                    </div>
                </div>
            </div>

            {/* Main Content */}
            <div className="flex-1 p-5 overflow-y-auto">
                <div className="max-w-full space-y-6">
                    {/* Timeline Charts */}
                    <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
                        {/* Viewer Timeline */}
                        <Card className="bg-muted/30 border-muted">
                            <CardHeader className="pb-4">
                                <CardTitle>Viewer Timeline</CardTitle>
                                <CardDescription>
                                    Viewer count tracked every 2 minutes
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="p-6">
                                <div className="w-full h-[300px]">
                                    <ChartContainer
                                        config={chartConfig}
                                        className="w-full h-full"
                                    >
                                        <ResponsiveContainer
                                            width="100%"
                                            height="100%"
                                        >
                                            <AreaChart
                                                data={viewer_counts_timeline}
                                                margin={{
                                                    top: 5,
                                                    right: 30,
                                                    left: 20,
                                                    bottom: 5,
                                                }}
                                            >
                                                <CartesianGrid
                                                    strokeDasharray="3 3"
                                                    stroke="hsl(var(--muted-foreground))"
                                                    opacity={0.2}
                                                />
                                                <XAxis
                                                    dataKey="time"
                                                    stroke="hsl(var(--muted-foreground))"
                                                    fontSize={12}
                                                    tick={{ fontSize: 12 }}
                                                    tickFormatter={formatTime}
                                                />
                                                <YAxis
                                                    stroke="hsl(var(--muted-foreground))"
                                                    fontSize={12}
                                                    tick={{ fontSize: 12 }}
                                                />
                                                <ChartTooltip
                                                    content={
                                                        <ChartTooltipContent />
                                                    }
                                                    labelFormatter={(value) =>
                                                        formatTime(
                                                            value as string,
                                                        )
                                                    }
                                                />
                                                <Area
                                                    type="monotone"
                                                    dataKey="count"
                                                    stroke="var(--color-viewers)"
                                                    fill="var(--color-viewers)"
                                                    fillOpacity={0.3}
                                                    strokeWidth={3}
                                                />
                                            </AreaChart>
                                        </ResponsiveContainer>
                                    </ChartContainer>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Message Timeline */}
                        <Card className="bg-muted/30 border-muted">
                            <CardHeader className="pb-4">
                                <CardTitle>Message Timeline</CardTitle>
                                <CardDescription>
                                    Message count tracked every 2 minutes
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="p-6">
                                <div className="w-full h-[300px]">
                                    <ChartContainer
                                        config={chartConfig}
                                        className="w-full h-full"
                                    >
                                        <ResponsiveContainer
                                            width="100%"
                                            height="100%"
                                        >
                                            <AreaChart
                                                data={message_counts_timeline}
                                                margin={{
                                                    top: 5,
                                                    right: 30,
                                                    left: 20,
                                                    bottom: 5,
                                                }}
                                            >
                                                <CartesianGrid
                                                    strokeDasharray="3 3"
                                                    stroke="hsl(var(--muted-foreground))"
                                                    opacity={0.2}
                                                />
                                                <XAxis
                                                    dataKey="time"
                                                    stroke="hsl(var(--muted-foreground))"
                                                    fontSize={12}
                                                    tick={{ fontSize: 12 }}
                                                    tickFormatter={formatTime}
                                                />
                                                <YAxis
                                                    stroke="hsl(var(--muted-foreground))"
                                                    fontSize={12}
                                                    tick={{ fontSize: 12 }}
                                                />
                                                <ChartTooltip
                                                    content={
                                                        <ChartTooltipContent />
                                                    }
                                                    labelFormatter={(value) =>
                                                        formatTime(
                                                            value as string,
                                                        )
                                                    }
                                                />
                                                <Area
                                                    type="monotone"
                                                    dataKey="count"
                                                    stroke="var(--color-messages)"
                                                    fill="var(--color-messages)"
                                                    fillOpacity={0.3}
                                                    strokeWidth={3}
                                                />
                                            </AreaChart>
                                        </ResponsiveContainer>
                                    </ChartContainer>
                                </div>
                            </CardContent>
                        </Card>
                    </div>

                    {/* Suspicious Chatters Section */}
                    <Card className="bg-muted/30 border-muted">
                        <CardHeader className="pb-4">
                            <CardTitle className="flex items-center space-x-2">
                                <AlertTriangle className="h-5 w-5 text-red-500" />
                                <span>Suspicious Chatters</span>
                            </CardTitle>
                            <CardDescription>
                                Users flagged for suspicious behavior during the
                                stream
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {suspiciousChatters.length === 0 ? (
                                <div className="text-center py-8">
                                    <p className="text-muted-foreground">
                                        No suspicious chatters detected.
                                    </p>
                                </div>
                            ) : (
                                <>
                                    <div className="overflow-x-auto">
                                        <Table>
                                            <TableHeader>
                                                <TableRow>
                                                    <TableHead className="min-w-[150px]">
                                                        Username
                                                    </TableHead>
                                                    <TableHead>
                                                        Reason
                                                    </TableHead>
                                                    <TableHead className="hidden md:table-cell">
                                                        Evidence
                                                    </TableHead>
                                                    <TableHead className="hidden lg:table-cell">
                                                        Issues
                                                    </TableHead>
                                                    <TableHead className="hidden lg:table-cell">
                                                        Timeframe
                                                    </TableHead>
                                                    <TableHead>
                                                        Flagged At
                                                    </TableHead>
                                                </TableRow>
                                            </TableHeader>
                                            <TableBody>
                                                {paginatedSuspiciousChatters.map(
                                                    (chatter) => (
                                                        <TableRow
                                                            key={
                                                                chatter.user_id
                                                            }
                                                            className="hover:bg-muted/20"
                                                        >
                                                            <TableCell className="py-2">
                                                                <div className="flex items-center space-x-2">
                                                                    <Flag className="h-3 w-3 text-red-500 flex-shrink-0" />
                                                                    <span className="font-medium text-sm">
                                                                        {
                                                                            chatter.username
                                                                        }
                                                                    </span>
                                                                </div>
                                                            </TableCell>
                                                            <TableCell className="py-2">
                                                                <Badge
                                                                    variant="outline"
                                                                    className="bg-red-500/10 text-red-500 border-red-500/20 text-xs"
                                                                >
                                                                    {
                                                                        chatter
                                                                            .potential_issues[0]
                                                                    }
                                                                </Badge>
                                                            </TableCell>
                                                            <TableCell className="hidden md:table-cell py-2">
                                                                <span className="text-xs text-muted-foreground">
                                                                    {
                                                                        chatter
                                                                            .example_messages[0]
                                                                    }
                                                                </span>
                                                            </TableCell>
                                                            <TableCell className="hidden lg:table-cell py-2">
                                                                <span className="font-semibold text-sm">
                                                                    {
                                                                        chatter
                                                                            .potential_issues
                                                                            .length
                                                                    }
                                                                </span>
                                                            </TableCell>
                                                            <TableCell className="hidden lg:table-cell py-2">
                                                                <span className="text-xs">
                                                                    {
                                                                        chatter
                                                                            .message_timestamps
                                                                            .length
                                                                    }{" "}
                                                                    messages
                                                                </span>
                                                            </TableCell>
                                                            {/* <TableCell className="py-2"> */}
                                                            {/*     <span className="text-xs font-mono"> */}
                                                            {/*         {formatTime( */}
                                                            {/*             chatter.message_timestamps[0].toString(), */}
                                                            {/*         )} */}
                                                            {/*     </span> */}
                                                            {/* </TableCell> */}
                                                        </TableRow>
                                                    ),
                                                )}
                                            </TableBody>
                                        </Table>
                                    </div>
                                    <div className="mt-4">
                                        <Pagination
                                            currentPage={suspiciousPage}
                                            totalPages={suspiciousTotalPages}
                                            onPageChange={setSuspiciousPage}
                                        />
                                    </div>
                                </>
                            )}
                        </CardContent>
                    </Card>

                    {/* Exact Duplicate Bursts Section */}
                    <Card className="bg-muted/30 border-muted">
                        <CardHeader className="pb-4">
                            <CardTitle className="flex items-center space-x-2">
                                <Copy className="h-5 w-5 text-primary" />
                                <span>Exact Duplicate Bursts</span>
                            </CardTitle>
                            <CardDescription>
                                Messages that appeared in identical bursts
                            </CardDescription>

                            <div className="flex items-center space-x-4 pt-4">
                                <Select
                                    value={exactDuplicateSort}
                                    onValueChange={setExactDuplicateSort}
                                >
                                    <SelectTrigger className="w-40">
                                        <SelectValue placeholder="Sort by" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="count">
                                            Total Count
                                        </SelectItem>
                                        <SelectItem value="bursts">
                                            Burst Count
                                        </SelectItem>
                                        <SelectItem value="users">
                                            User Count
                                        </SelectItem>
                                        <SelectItem value="avgSize">
                                            Avg Burst Size
                                        </SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <div className="overflow-x-auto">
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead className="min-w-[200px]">
                                                Message Content
                                            </TableHead>
                                            <TableHead>Total Count</TableHead>
                                            <TableHead className="hidden md:table-cell">
                                                Bursts
                                            </TableHead>
                                            <TableHead className="hidden lg:table-cell">
                                                Top User
                                            </TableHead>
                                            <TableHead className="hidden lg:table-cell">
                                                Unique Users
                                            </TableHead>
                                            <TableHead>
                                                Avg Burst Size
                                            </TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {paginatedExactDuplicates.map(
                                            (message, index) => (
                                                <TableRow
                                                    key={message.username}
                                                    className="hover:bg-muted/20"
                                                >
                                                    <TableCell className="py-2">
                                                        <div className="flex items-center space-x-2">
                                                            <span className="text-xs font-medium">
                                                                #
                                                                {(exactDuplicatePage -
                                                                    1) *
                                                                    itemsPerPage +
                                                                    index +
                                                                    1}
                                                            </span>
                                                            <div className="bg-muted/40 rounded px-2 py-1 font-mono text-xs">
                                                                {
                                                                    message.content
                                                                }
                                                            </div>
                                                        </div>
                                                    </TableCell>
                                                    <TableCell className="py-2">
                                                        <span className="font-semibold text-primary text-sm">
                                                            {formatNumber(
                                                                message.count,
                                                            )}
                                                        </span>
                                                    </TableCell>
                                                    <TableCell className="hidden md:table-cell py-2">
                                                        <Badge
                                                            variant="outline"
                                                            className="bg-yellow-500/10 text-yellow-500 border-yellow-500/20 text-xs"
                                                        >
                                                            {message.count}{" "}
                                                            bursts
                                                        </Badge>
                                                    </TableCell>
                                                    <TableCell className="hidden lg:table-cell py-2">
                                                        <span className="text-xs text-muted-foreground">
                                                            {message.username}
                                                        </span>
                                                    </TableCell>
                                                    {/* <TableCell className="hidden lg:table-cell py-2"> */}
                                                    {/*     <span className="text-xs"> */}
                                                    {/*         {message.userCount}{" "} */}
                                                    {/*         users */}
                                                    {/*     </span> */}
                                                    {/* </TableCell> */}
                                                    {/* <TableCell className="py-2"> */}
                                                    {/*     <span className="font-semibold text-sm"> */}
                                                    {/*         {message.avgBurstSize.toFixed( */}
                                                    {/*             1, */}
                                                    {/*         )} */}
                                                    {/*     </span> */}
                                                    {/* </TableCell> */}
                                                </TableRow>
                                            ),
                                        )}
                                    </TableBody>
                                </Table>
                            </div>
                            <div className="mt-4">
                                <Pagination
                                    currentPage={exactDuplicatePage}
                                    totalPages={exactDuplicateTotalPages}
                                    onPageChange={setExactDuplicatePage}
                                />
                            </div>
                        </CardContent>
                    </Card>

                    {/* Similar Message Bursts Section */}
                    <Card className="bg-muted/30 border-muted">
                        <CardHeader className="pb-4">
                            <CardTitle className="flex items-center space-x-2">
                                <MessageSquare className="h-5 w-5 text-blue-400" />
                                <span>Similar Message Bursts</span>
                            </CardTitle>
                            <CardDescription>
                                Messages with similar patterns that appeared in
                                bursts
                            </CardDescription>

                            <div className="flex items-center space-x-4 pt-4">
                                <Select
                                    value={similarMessageSort}
                                    onValueChange={setSimilarMessageSort}
                                >
                                    <SelectTrigger className="w-40">
                                        <SelectValue placeholder="Sort by" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="count">
                                            Total Count
                                        </SelectItem>
                                        <SelectItem value="bursts">
                                            Burst Count
                                        </SelectItem>
                                        <SelectItem value="users">
                                            User Count
                                        </SelectItem>
                                        <SelectItem value="similarity">
                                            Similarity %
                                        </SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <div className="overflow-x-auto">
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead className="min-w-[200px]">
                                                Pattern
                                            </TableHead>
                                            <TableHead>Total Count</TableHead>
                                            <TableHead className="hidden md:table-cell">
                                                Bursts
                                            </TableHead>
                                            <TableHead className="hidden lg:table-cell">
                                                Top User
                                            </TableHead>
                                            <TableHead className="hidden lg:table-cell">
                                                Unique Users
                                            </TableHead>
                                            <TableHead>Similarity</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {paginatedSimilarMessages.map(
                                            (message, index) => (
                                                <TableRow
                                                    key={message.username}
                                                    className="hover:bg-muted/20"
                                                >
                                                    <TableCell className="py-2">
                                                        <div className="space-y-1">
                                                            <div className="flex items-center space-x-2">
                                                                <span className="text-xs font-medium">
                                                                    #
                                                                    {(similarMessagePage -
                                                                        1) *
                                                                        itemsPerPage +
                                                                        index +
                                                                        1}
                                                                </span>
                                                                <div className="bg-muted/40 rounded px-2 py-1 font-mono text-xs">
                                                                    {
                                                                        message.pattern
                                                                    }
                                                                </div>
                                                            </div>
                                                            {/* <div className="text-xs text-muted-foreground ml-6"> */}
                                                            {/*     { */}
                                                            {/*         message */}
                                                            {/*             .variations */}
                                                            {/*             .length */}
                                                            {/*     } */}
                                                            {/*     variations */}
                                                            {/* </div> */}
                                                        </div>
                                                    </TableCell>
                                                    <TableCell className="py-2">
                                                        <span className="font-semibold text-primary text-sm">
                                                            {formatNumber(
                                                                message.count,
                                                            )}
                                                        </span>
                                                    </TableCell>
                                                    <TableCell className="hidden md:table-cell py-2">
                                                        <Badge
                                                            variant="outline"
                                                            className="bg-blue-500/10 text-blue-500 border-blue-500/20 text-xs"
                                                        >
                                                            {message.count}{" "}
                                                            bursts
                                                        </Badge>
                                                    </TableCell>
                                                    <TableCell className="hidden lg:table-cell py-2">
                                                        <span className="text-xs text-muted-foreground">
                                                            {message.username}
                                                        </span>
                                                    </TableCell>
                                                    <TableCell className="hidden lg:table-cell py-2">
                                                        <span className="text-xs">
                                                            {message.count}{" "}
                                                            users
                                                        </span>
                                                    </TableCell>
                                                    <TableCell className="py-2">
                                                        <Badge
                                                            variant="outline"
                                                            className="bg-green-500/10 text-green-500 border-green-500/20 text-xs"
                                                        >
                                                            {/* {message.similarity.toFixed( */}
                                                            {/*     1, */}
                                                            {/* )} */}%
                                                        </Badge>
                                                    </TableCell>
                                                </TableRow>
                                            ),
                                        )}
                                    </TableBody>
                                </Table>
                            </div>
                            <div className="mt-4">
                                <Pagination
                                    currentPage={similarMessagePage}
                                    totalPages={similarMessageTotalPages}
                                    onPageChange={setSimilarMessagePage}
                                />
                            </div>
                        </CardContent>
                    </Card>

                    {/* Message Count Heatmap */}
                    <Card className="bg-muted/30 border-muted">
                        <CardHeader className="pb-4">
                            <CardTitle className="flex items-center space-x-2">
                                <TrendingUp className="h-5 w-5 text-orange-400" />
                                <span>Message Activity Heatmap</span>
                            </CardTitle>
                            <CardDescription>
                                Message count intensity over 3 hours (2-minute
                                intervals)
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-6">
                                {/* Legend */}
                                <div className="flex items-center justify-between text-sm">
                                    <span className="text-muted-foreground">
                                        Activity Level:
                                    </span>
                                    <div className="flex items-center space-x-3">
                                        <span className="text-xs text-muted-foreground">
                                            Low
                                        </span>
                                        <div className="flex space-x-1">
                                            <div className="w-4 h-4 bg-muted/30 border border-muted rounded-sm"></div>
                                            <div className="w-4 h-4 bg-primary/20 border border-primary/15 rounded-sm"></div>
                                            <div className="w-4 h-4 bg-primary/40 border border-primary/20 rounded-sm"></div>
                                            <div className="w-4 h-4 bg-primary/60 border border-primary/25 rounded-sm"></div>
                                            <div className="w-4 h-4 bg-primary/80 border border-primary/30 rounded-sm"></div>
                                            <div className="w-4 h-4 bg-primary border border-primary/40 rounded-sm"></div>
                                        </div>
                                        <span className="text-xs text-muted-foreground">
                                            High
                                        </span>
                                    </div>
                                </div>

                                {/* Heatmap Grid - Dynamically rendered for stream duration */}
                                {heatmapHours.length > 0 ? (
                                    <div className="space-y-3">
                                        {heatmapHours.map((hour) => (
                                            <div
                                                key={hour}
                                                className="space-y-2"
                                            >
                                                <div className="flex items-center space-x-2">
                                                    <span className="text-sm font-medium text-muted-foreground w-12">
                                                        {String(hour).padStart(
                                                            2,
                                                            "0",
                                                        )}
                                                        :00
                                                    </span>
                                                    <div className="flex-1 h-px bg-muted"></div>
                                                </div>
                                                <div className="ml-14">
                                                    {/* We need 30 cells per hour (60 mins / 2 min interval = 30) */}
                                                    <div className="grid grid-cols-30 gap-1 w-full">
                                                        {Array.from({
                                                            length: 30,
                                                        }).map((_, i) => {
                                                            const minute =
                                                                i * 2;
                                                            const dataPoint =
                                                                heatmapData.find(
                                                                    (d) =>
                                                                        d.hour ===
                                                                            hour &&
                                                                        d.minute ===
                                                                            minute,
                                                                );
                                                            const count =
                                                                dataPoint
                                                                    ? dataPoint.count
                                                                    : 0; // Use 0 if no data
                                                            return (
                                                                <div
                                                                    key={`${hour}-${minute}`}
                                                                    className={`aspect-square rounded-sm border transition-all duration-200 hover:scale-110 relative group cursor-pointer ${getHeatmapColor(count, maxHeatmapCount)}`}
                                                                    title={`${String(hour).padStart(2, "0")}:${String(minute).padStart(2, "0")} - ${formatNumber(count)} messages`}
                                                                >
                                                                    {/* Tooltip on hover */}
                                                                    <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-black/90 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity duration-200 whitespace-nowrap z-20 pointer-events-none">
                                                                        {String(
                                                                            hour,
                                                                        ).padStart(
                                                                            2,
                                                                            "0",
                                                                        )}
                                                                        :
                                                                        {String(
                                                                            minute,
                                                                        ).padStart(
                                                                            2,
                                                                            "0",
                                                                        )}{" "}
                                                                        -{" "}
                                                                        {formatNumber(
                                                                            count,
                                                                        )}{" "}
                                                                        msgs
                                                                        <div className="absolute top-full left-1/2 transform -translate-x-1/2 border-2 border-transparent border-t-black/90"></div>
                                                                    </div>
                                                                </div>
                                                            );
                                                        })}
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="text-center py-8">
                                        <p className="text-muted-foreground">
                                            No message activity data available
                                            for heatmap.
                                        </p>
                                    </div>
                                )}

                                {/* Time Markers (Simplified to show start/mid/end for heatmap's timeframe) */}
                                <div className="flex justify-between text-xs text-muted-foreground ml-14">
                                    <span>
                                        {formatTime(
                                            data.report_start_time.toString(),
                                        )}
                                    </span>
                                    {/* {duration_minutes > 120 && ( // Only show if stream is longer than 2 hours */}
                                    {/*     <span> */}
                                    {/*         {formatTime( */}
                                    {/*             new Date( */}
                                    {/*                 data.report_start_time.getTime() + */}
                                    {/*                     (duration_minutes / 2) * */}
                                    {/*                         60 * */}
                                    {/*                         1000, */}
                                    {/*             ), */}
                                    {/*         )} */}
                                    {/*     </span> */}
                                    {/* )} */}
                                    <span>
                                        {formatTime(
                                            data.report_end_time.toString(),
                                        )}
                                    </span>
                                </div>

                                {/* Enhanced Stats */}
                                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-6 border-t border-muted">
                                    <div className="text-center">
                                        <div className="text-xl font-bold text-primary">
                                            {formatNumber(maxHeatmapCount)}
                                        </div>
                                        <div className="text-xs text-muted-foreground">
                                            Peak Activity
                                        </div>
                                    </div>
                                    <div className="text-center">
                                        <div className="text-xl font-bold text-yellow-400">
                                            {formatNumber(
                                                Math.round(
                                                    mockHeatmapData.reduce(
                                                        (sum, d) =>
                                                            sum + d.count,
                                                        0,
                                                    ) / mockHeatmapData.length,
                                                ),
                                            )}
                                        </div>
                                        <div className="text-xs text-muted-foreground">
                                            Avg Activity
                                        </div>
                                    </div>
                                    <div className="text-center">
                                        <div className="text-xl font-bold text-green-400">
                                            {formatNumber(
                                                mockHeatmapData.reduce(
                                                    (sum, d) => sum + d.count,
                                                    0,
                                                ),
                                            )}
                                        </div>
                                        <div className="text-xs text-muted-foreground">
                                            Total Messages
                                        </div>
                                    </div>
                                    <div className="text-center">
                                        <div className="text-xl font-bold text-blue-400">
                                            3h
                                        </div>
                                        <div className="text-xs text-muted-foreground">
                                            Duration
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
