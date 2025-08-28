export interface APIResponse {
    channel_id: number;
    username: string;
    verified: boolean;
    is_banned: boolean;
    vod_enabled: boolean;
    is_affiliate: boolean;
    subscription_enabled: boolean;
    followers_count: {
        time: Date;
        count: number;
    }[];
    livestreams: Livestream[] | null;
    twitter: string;
    profile_pic: string;
}

export interface Livestream {
    livestream_id: number;
    title: number;
    report_start_time: Date;
    report_end_time: Date;
    duration_minutes: number;
    average_viewers: number;
    peak_viewers: number;
    lowest_viewers: number;
    engagement: number;
    hours_watched: number;
    total_messages: number;
    unique_chatters: number;
    messages_from_apps: number;
    viewer_counts_timeline: {
        time: Date;
        count: number;
    }[];
    message_counts_timeline: {
        time: Date;
        count: number;
    }[];
    created_at: Date;
    spam_report: SpamReport;
}

export interface SpamReport {
    messages_with_emotes: number;
    messages_multiple_emotes_only: number;
    duplicate_messages_count: number;
    repetitive_phrases_count: number;
    exact_duplicate_bursts: {
        count: number;
        content?: string;
        username: string;
        timestamps: Date[];
        pattern?: string;
    }[];

    similar_message_bursts: {
        count: number;
        content?: string;
        username: string;
        timestamps: Date[];
        pattern?: string;
    }[];
    suspicious_chatters: {
        user_id: number;
        username: string;
        example_messages: string[];
        potential_issues: string[];
        message_timestamps: Date[];
    }[];
}

export interface AllLivestreams {
    ChannelID: number;
    LivestreamID: number;
    Slug: string;
    StartTime: Date;
    SessionTitle: string;
    ViewerCount: number;
    LivestreamCreatedAt: Date;
    Tags: string;
    IsLive: boolean;
    Duration: number;
    LangISO: string;
    CreatedAt: Date;
}
