import { useState } from "react";
import type React from "react";
import { Search, Play, Loader2, Eye, AlertCircle } from "lucide-react";
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
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "../components/ui/table";
import { Link } from "react-router";
import { useQuery } from "@tanstack/react-query";
import type { AllLivestreams } from "../types/types";

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

const fetchLivestreams = async (): Promise<AllLivestreams[]> => {
    const response = await fetch(`http://localhost:80/api/livestreams`);
    if (!response.ok) {
        throw new Error(`Failed to fetch livestreams: ${response.statusText}`);
    }
    const data = await response.json();
    return data.map((stream: any) => ({
        ...stream,
        CreatedAt: new Date(stream.CreatedAt),
    }));
};

const useAllLivestreamData = () => {
    return useQuery({
        queryKey: ["all-livestreams"],
        queryFn: fetchLivestreams,
        staleTime: 5 * 60 * 1000, // 5 minutes
        retry: 3,
        retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    });
};

export default function AdminPage() {
    const [searchTerm, setSearchTerm] = useState("");
    const [searchResults, setSearchResults] = useState<AllLivestreams[]>([]);
    const [showSearchResults, setShowSearchResults] = useState(false);

    const {
        data: allLivestreams,
        isLoading,
        isError,
        error,
        refetch,
    } = useAllLivestreamData();

    const handleSearch = (e?: React.FormEvent) => {
        if (e) e.preventDefault();

        if (!searchTerm.trim()) {
            setShowSearchResults(false);
            return;
        }

        if (allLivestreams) {
            const lowercasedTerm = searchTerm.toLowerCase();
            const results = allLivestreams.filter(
                (stream) =>
                    stream.SessionTitle.toLowerCase().includes(
                        lowercasedTerm,
                    ) ||
                    stream.ChannelID.toString()
                        .toLowerCase()
                        .includes(lowercasedTerm) ||
                    stream.LivestreamID.toString()
                        .toLowerCase()
                        .includes(lowercasedTerm),
            );
            setSearchResults(results);
        }
        setShowSearchResults(true);
    };

    const handleProcess = async (id: number) => {
        try {
            const response = await fetch(
                "http://localhost:80/api/process_livestream_report",
                {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify({
                        livestream_id: Number(id),
                    }),
                },
            );

            if (response.ok) {
                // TODO: Add a ui indication of success.
                // toast.success(`Successfully started processing: ${title}`);
                console.log(`Successfully started processing: ${id}`);
            } else {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
        } catch (error) {
            console.error("Processing error:", error);
            // toast.error(`Failed to process: ${title}. Please try again.`);
        } finally {
        }
    };

    const StreamRow = ({ stream }: { stream: AllLivestreams }) => (
        <TableRow key={stream.LivestreamID} className="hover:bg-muted/20">
            <TableCell className="py-3">
                <div className="space-y-1">
                    <div className="font-medium text-sm">
                        {stream.SessionTitle}
                    </div>
                    <div className="text-xs text-muted-foreground">
                        ID: {stream.LivestreamID}
                    </div>
                </div>
            </TableCell>
            <TableCell className="hidden lg:table-cell py-3">
                <div className="text-sm">
                    {stream.CreatedAt.toLocaleDateString()}
                </div>
            </TableCell>
            <TableCell className="py-3">
                <div className="flex items-center space-x-2">
                    <Button
                        size="sm"
                        onClick={() => handleProcess(stream.LivestreamID)}
                        className="h-8 px-3"
                    >
                        <Play className="h-3 w-3 mr-1" />
                        Process
                    </Button>
                    <Link to={`/stream/${stream.LivestreamID}`}>
                        <Button
                            variant="outline"
                            size="sm"
                            className="h-8 px-3 bg-transparent"
                        >
                            <Eye className="h-3 w-3 mr-1" />
                            View
                        </Button>
                    </Link>
                </div>
            </TableCell>
        </TableRow>
    );

    if (isLoading) {
        return (
            <div className="min-h-screen bg-muted/20 flex items-center justify-center">
                <LoadingSpinner />
            </div>
        );
    }

    if (isError || !allLivestreams) {
        return (
            <div className="min-h-screen bg-muted/20 flex items-center justify-center">
                <ErrorDisplay
                    error={error as Error}
                    onRetry={() => refetch()}
                />
            </div>
        );
    }

    const stats = {
        total: allLivestreams.length,
    };

    return (
        <div className="min-h-screen bg-muted/20 p-6">
            <div className="max-w-7xl mx-auto space-y-6">
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold">Admin Dashboard</h1>
                        <p className="text-muted-foreground">
                            Manage and process livestream data
                        </p>
                    </div>
                    <Link to="/admin">
                        <Button variant="outline">Back to Dashboard</Button>
                    </Link>
                </div>

                <div className="grid justify-around grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">
                                Total Streams
                            </CardTitle>
                            <Eye className="h-4 w-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">
                                {stats.total}
                            </div>
                        </CardContent>
                    </Card>
                </div>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center space-x-2">
                            <Search className="h-5 w-5 text-primary" />
                            <span>Search Livestreams</span>
                        </CardTitle>
                        <CardDescription>
                            Search by title, streamer, or livestream ID.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form
                            onSubmit={handleSearch}
                            className="flex space-x-4"
                        >
                            <Input
                                placeholder="Enter search term..."
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                                className="bg-background/50"
                            />
                            <Button type="submit">
                                <Search className="h-4 w-4 mr-2" />
                                Search
                            </Button>
                        </form>
                    </CardContent>
                </Card>

                {showSearchResults && (
                    <Card>
                        <CardHeader>
                            <CardTitle>Search Results</CardTitle>
                            <CardDescription>
                                Found {searchResults.length} livestream(s)
                                matching "{searchTerm}"
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {searchResults.length > 0 ? (
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead className="min-w-[250px]">
                                                Title
                                            </TableHead>
                                            {/* <TableHead>Streamer</TableHead> */}
                                            <TableHead className="hidden lg:table-cell">
                                                Date
                                            </TableHead>
                                            <TableHead>Actions</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {searchResults.map((stream) => (
                                            <StreamRow
                                                key={stream.LivestreamID}
                                                stream={stream}
                                            />
                                        ))}
                                    </TableBody>
                                </Table>
                            ) : (
                                <div className="text-center py-8 text-muted-foreground">
                                    No livestreams found.
                                </div>
                            )}
                        </CardContent>
                    </Card>
                )}

                <Card>
                    <CardHeader>
                        <CardTitle>All Livestreams</CardTitle>
                        <CardDescription>
                            A complete list of all {allLivestreams.length}{" "}
                            recorded livestreams.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="overflow-x-auto">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead className="min-w-[250px]">
                                            Title
                                        </TableHead>
                                        {/* <TableHead>Streamer</TableHead> */}
                                        <TableHead className="hidden lg:table-cell">
                                            Date
                                        </TableHead>
                                        <TableHead>Actions</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {allLivestreams.map((stream) => (
                                        <StreamRow
                                            key={stream.LivestreamID}
                                            stream={stream}
                                        />
                                    ))}
                                </TableBody>
                            </Table>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
