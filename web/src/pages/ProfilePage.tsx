import { useState, useMemo } from "react";
import {
    Star,
    Search,
    Loader2,
    AlertCircle,
    MoveUpRight,
    Calendar as CalendarIcon,
} from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { addDays, format } from "date-fns";
import { type DateRange } from "react-day-picker";

import { Avatar, AvatarFallback, AvatarImage } from "../components/ui/avatar";
import { Button } from "../components/ui/button";
import { Calendar } from "../components/ui/calendar";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "../components/ui/popover";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "../components/ui/card";
import { Badge } from "../components/ui/badge";
import { Input } from "../components/ui/input";
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
    Line,
    LineChart,
    XAxis,
    YAxis,
    CartesianGrid,
    ResponsiveContainer,
} from "recharts";

import { useParams, Link } from "react-router";
import type { APIResponse } from "../types/types";
import { cn } from "../lib/utils";

const chartConfig = {
    viewers: {
        label: "Average Viewers",
        color: "oklch(0.6776 0.1481 238.1)",
    },
    followers: {
        label: "Total Followers",
        color: "oklch(0.6776 0.1481 238.1)",
    },
} satisfies ChartConfig;

const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return num.toLocaleString();
};

const getEngagementColor = (engagement: number) => {
    if (engagement >= 9) return "text-green-500";
    if (engagement >= 8) return "text-primary";
    if (engagement >= 7) return "text-yellow-500";
    return "text-red-500";
};

const formatXAxis = (tickItem: string, aggregationLevel: "hour" | "day") => {
    if (aggregationLevel === "hour") {
        return format(new Date(tickItem), "h:mm a");
    }
    return format(new Date(tickItem), "MMM d");
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
                Failed to load profile data
            </h3>
            <p className="text-muted-foreground mb-4">{error.message}</p>
            <Button onClick={onRetry} variant="outline">
                Try Again
            </Button>
        </div>
    </div>
);

const fetchProfileData = async (username: string): Promise<APIResponse> => {
    const response = await fetch(`http://localhost:80/api/profile/${username}`);
    if (!response.ok) {
        throw new Error(`Failed to fetch profile data: ${response.statusText}`);
    }
    return response.json();
};

const useProfileData = (username: string | undefined) => {
    return useQuery({
        queryKey: ["profile", username],
        queryFn: () => fetchProfileData(username!),
        enabled: !!username,
        staleTime: 5 * 60 * 1000,
        retry: 3,
        retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    });
};

const simplifyTimeSeriesData = (
    data: { time: Date | string; count: number }[],
    aggregationLevel: "hour" | "day",
): { time: string; count: number }[] => {
    if (!data || data.length === 0) return [];

    const groupedData: { [key: string]: number[] } = {};

    data.forEach((item) => {
        const date = new Date(item.time);
        let key: string;

        if (aggregationLevel === "hour") {
            date.setMinutes(0, 0, 0);
            key = date.toISOString();
        } else {
            date.setHours(0, 0, 0, 0);
            key = date.toISOString().split("T")[0];
        }

        if (!groupedData[key]) {
            groupedData[key] = [];
        }
        groupedData[key].push(item.count);
    });

    const simplifiedData = Object.entries(groupedData).map(([time, counts]) => {
        const finalCount =
            aggregationLevel === "day"
                ? counts[counts.length - 1]
                : Math.round(
                      counts.reduce((sum, count) => sum + count, 0) /
                          counts.length,
                  );
        const timeKey =
            aggregationLevel === "day" ? new Date(time).toISOString() : time;
        return { time: timeKey, count: finalCount };
    });

    return simplifiedData.sort(
        (a, b) => new Date(a.time).getTime() - new Date(b.time).getTime(),
    );
};

export default function Dashboard() {
    const { username } = useParams<{ username: string }>();
    const { data, isLoading, isError, error, refetch } =
        useProfileData(username);

    const [searchTerm, setSearchTerm] = useState("");
    const [sortBy, setSortBy] = useState("average_viewers");

    const [dateRange, setDateRange] = useState<DateRange | undefined>({
        from: addDays(new Date(), -30),
        to: new Date(),
    });

    const filteredStreams = useMemo(() => {
        if (!data?.livestreams) return [];
        const filtered = data.livestreams.filter((stream) =>
            stream.livestream_id.toString().includes(searchTerm.toLowerCase()),
        );
        filtered.sort((a, b) => {
            switch (sortBy) {
                case "average_viewers":
                    return b.average_viewers - a.average_viewers;
                case "peak_viewers":
                    return b.peak_viewers - a.peak_viewers;
                case "hours_watched":
                    return b.hours_watched - a.hours_watched;
                case "engagement":
                    return b.engagement - a.engagement;
                case "messages":
                    return b.total_messages - a.total_messages;
                default:
                    return b.average_viewers - a.average_viewers;
            }
        });
        return filtered;
    }, [searchTerm, sortBy, data?.livestreams]);

    const processedChartData = useMemo(() => {
        if (!data?.followers_count) return { data: [], level: "day" };

        const filteredData = data.followers_count.filter((item) => {
            const itemDate = new Date(item.time);
            if (!dateRange?.from) return true;
            const fromDate = new Date(dateRange.from);
            fromDate.setHours(0, 0, 0, 0);
            const toDate = dateRange.to ? new Date(dateRange.to) : new Date();
            toDate.setHours(23, 59, 59, 999);
            return itemDate >= fromDate && itemDate <= toDate;
        });

        const dayDifference =
            dateRange?.from && dateRange?.to
                ? (dateRange.to.getTime() - dateRange.from.getTime()) /
                  (1000 * 3600 * 24)
                : 0;
        const aggregationLevel = dayDifference <= 3 ? "hour" : "day";

        return {
            data: simplifyTimeSeriesData(filteredData, aggregationLevel),
            level: aggregationLevel,
        };
    }, [data?.followers_count, dateRange]);

    if (isLoading) {
        return (
            <div className="min-h-screen bg-muted/20 flex items-center justify-center">
                <LoadingSpinner />
            </div>
        );
    }

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

    const { profile_pic, username: data_username } = data;

    return (
        <div className="min-h-screen bg-muted/20 flex flex-col lg:flex-row">
            {/* Side Panel */}
            <div className="w-full lg:w-96 bg-muted/40 border-b lg:border-b-0 lg:border-r border-muted p-5 overflow-y-auto">
                {/* Profile Header */}
                <div className="mb-6">
                    <div className="flex items-center justify-between mb-6">
                        <div className="flex items-center space-x-4">
                            <Avatar className="h-16 w-16 border-4 border-primary/20">
                                <AvatarImage
                                    src={profile_pic || "/placeholder.svg"}
                                    alt={data_username}
                                />
                                <AvatarFallback className="text-xl font-bold">
                                    XQ
                                </AvatarFallback>
                            </Avatar>
                            <div>
                                <h1 className="text-2xl font-bold">
                                    {data_username}
                                </h1>
                                <p className="text-muted-foreground">
                                    @{data_username}
                                </p>
                            </div>
                        </div>
                    </div>
                    <div className="flex items-center justify-between mb-4">
                        {data.subscription_enabled ? (
                            <Badge
                                variant="secondary"
                                className="bg-primary/10 text-primary border-primary/20"
                            >
                                Partner
                            </Badge>
                        ) : (
                            <Badge
                                variant="secondary"
                                className="bg-primary/10 text-primary border-primary/20"
                            >
                                Affliate
                            </Badge>
                        )}
                        <Badge
                            variant="outline"
                            className="bg-primary text-primary-foreground border-primary"
                        >
                            <Star className="h-3 w-3 mr-1" />
                            Verified
                        </Badge>
                    </div>
                </div>
            </div>

            {/* Main Content */}
            <div className="flex-1 p-5 overflow-y-auto">
                <div className="max-w-full space-y-6">
                    {/* Charts Section */}
                    <div className="w-full">
                        <div className="space-y-6">
                            <div className="grid grid-cols-1 gap-6">
                                <Card className="bg-muted/30 border-muted">
                                    <CardHeader className="pb-4">
                                        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
                                            <div>
                                                <CardTitle>
                                                    Followers Count
                                                </CardTitle>
                                                <CardDescription>
                                                    Follower trends over the
                                                    selected period.
                                                </CardDescription>
                                            </div>
                                            <div className="flex items-center gap-2 mt-4 sm:mt-0">
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() =>
                                                        setDateRange({
                                                            from: addDays(
                                                                new Date(),
                                                                -7,
                                                            ),
                                                            to: new Date(),
                                                        })
                                                    }
                                                >
                                                    Last 7 Days
                                                </Button>
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() =>
                                                        setDateRange({
                                                            from: addDays(
                                                                new Date(),
                                                                -30,
                                                            ),
                                                            to: new Date(),
                                                        })
                                                    }
                                                >
                                                    Last 30 Days
                                                </Button>
                                                <Popover>
                                                    <PopoverTrigger asChild>
                                                        <Button
                                                            id="date"
                                                            variant={"outline"}
                                                            size="sm"
                                                            className={cn(
                                                                "w-[240px] justify-start text-left font-normal",
                                                                !dateRange &&
                                                                    "text-muted-foreground",
                                                            )}
                                                        >
                                                            <CalendarIcon className="mr-2 h-4 w-4" />
                                                            {dateRange?.from ? (
                                                                dateRange.to ? (
                                                                    <>
                                                                        {format(
                                                                            dateRange.from,
                                                                            "LLL dd, y",
                                                                        )}{" "}
                                                                        -{" "}
                                                                        {format(
                                                                            dateRange.to,
                                                                            "LLL dd, y",
                                                                        )}
                                                                    </>
                                                                ) : (
                                                                    format(
                                                                        dateRange.from,
                                                                        "LLL dd, y",
                                                                    )
                                                                )
                                                            ) : (
                                                                <span>
                                                                    Pick a date
                                                                </span>
                                                            )}
                                                        </Button>
                                                    </PopoverTrigger>
                                                    <PopoverContent
                                                        className="w-auto p-0"
                                                        align="end"
                                                    >
                                                        <Calendar
                                                            initialFocus
                                                            mode="range"
                                                            defaultMonth={
                                                                dateRange?.from
                                                            }
                                                            selected={dateRange}
                                                            onSelect={
                                                                setDateRange
                                                            }
                                                            numberOfMonths={2}
                                                        />
                                                    </PopoverContent>
                                                </Popover>
                                            </div>
                                        </div>
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
                                                    <LineChart
                                                        data={
                                                            processedChartData.data
                                                        }
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
                                                            tickFormatter={(
                                                                value,
                                                            ) =>
                                                                formatXAxis(
                                                                    value,
                                                                    processedChartData.level as
                                                                        | "hour"
                                                                        | "day",
                                                                )
                                                            }
                                                        />
                                                        <YAxis
                                                            stroke="hsl(var(--muted-foreground))"
                                                            fontSize={12}
                                                            tickFormatter={
                                                                formatNumber
                                                            }
                                                        />
                                                        <ChartTooltip
                                                            cursor={false}
                                                            content={
                                                                <ChartTooltipContent
                                                                    indicator="line"
                                                                    hideLabel
                                                                />
                                                            }
                                                        />
                                                        <Line
                                                            type="monotone"
                                                            dataKey="count"
                                                            name="followers"
                                                            stroke="var(--color-followers)"
                                                            strokeWidth={2}
                                                            dot={{
                                                                r: 2,
                                                                fill: "var(--color-followers)",
                                                            }}
                                                            activeDot={{
                                                                r: 6,
                                                                strokeWidth: 1,
                                                            }}
                                                        />
                                                    </LineChart>
                                                </ResponsiveContainer>
                                            </ChartContainer>
                                        </div>
                                    </CardContent>
                                </Card>
                            </div>
                        </div>
                    </div>

                    {/* Streams Table Section */}
                    <div className="w-full">
                        <Card className="bg-muted/30 border-muted">
                            <CardHeader className="pb-4">
                                <CardTitle>Recent Streams</CardTitle>
                                <CardDescription>
                                    Latest streaming sessions with performance
                                    metrics
                                </CardDescription>
                                <div className="flex flex-col sm:flex-row gap-4 pt-4">
                                    <div className="relative flex-1">
                                        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                                        <Input
                                            placeholder="Search streams..."
                                            value={searchTerm}
                                            onChange={(e) =>
                                                setSearchTerm(e.target.value)
                                            }
                                            className="pl-10"
                                        />
                                    </div>
                                    <div className="flex gap-2">
                                        <Select
                                            value={sortBy}
                                            onValueChange={setSortBy}
                                        >
                                            <SelectTrigger className="w-40">
                                                <SelectValue placeholder="Sort by" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="average_viewers">
                                                    Avg Viewers
                                                </SelectItem>
                                                <SelectItem value="peak_viewers">
                                                    Peak Viewers
                                                </SelectItem>
                                                <SelectItem value="hours_watched">
                                                    Hours Watched
                                                </SelectItem>
                                                <SelectItem value="engagement">
                                                    Engagement
                                                </SelectItem>
                                                <SelectItem value="messages">
                                                    Messages
                                                </SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="overflow-x-auto">
                                    <Table>
                                        <TableHeader>
                                            <TableRow>
                                                <TableHead className="min-w-[200px]">
                                                    Stream title
                                                </TableHead>
                                                <TableHead className="hidden md:table-cell">
                                                    Stream id
                                                </TableHead>
                                                <TableHead>
                                                    Avg Viewers
                                                </TableHead>
                                                <TableHead>
                                                    Engagement
                                                </TableHead>
                                            </TableRow>
                                        </TableHeader>
                                        <TableBody>
                                            {filteredStreams?.map((stream) => (
                                                <TableRow
                                                    key={stream.livestream_id}
                                                    className="hover:bg-muted/20"
                                                >
                                                    <TableCell>
                                                        <Link
                                                            to={`/stream/${stream.livestream_id}`}
                                                            className="font-medium flex items-center gap-2 text-sm leading-tight underline underline-offset-4"
                                                        >
                                                            <MoveUpRight
                                                                height={14}
                                                            />
                                                            <span>
                                                                {stream.title}
                                                            </span>
                                                        </Link>
                                                    </TableCell>
                                                    <TableCell className="hidden md:table-cell text-sm text-muted-foreground">
                                                        {stream.livestream_id}
                                                    </TableCell>
                                                    <TableCell>
                                                        <span className="font-semibold text-primary">
                                                            {formatNumber(
                                                                stream.average_viewers,
                                                            )}
                                                        </span>
                                                    </TableCell>
                                                    <TableCell>
                                                        <span
                                                            className={`font-semibold ${getEngagementColor(
                                                                stream.engagement,
                                                            )}`}
                                                        >
                                                            {stream.engagement.toFixed(
                                                                1,
                                                            )}
                                                        </span>
                                                    </TableCell>
                                                </TableRow>
                                            ))}
                                        </TableBody>
                                    </Table>
                                </div>
                                {filteredStreams?.length === 0 && (
                                    <div className="text-center py-8">
                                        <p className="text-muted-foreground">
                                            No streams found matching your
                                            criteria.
                                        </p>
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        </div>
    );
}
