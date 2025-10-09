"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
    Play,
    Save,
    Settings,
    History,
    Globe,
    Key,
    Code,
    Send,
    Copy,
    Check,
    AlertCircle,
    CheckCircle,
    Clock
} from "lucide-react";
import { toast } from "sonner";

interface APIEndpoint {
    method: string;
    path: string;
    description: string;
    parameters: Array<{
        name: string;
        type: string;
        required: boolean;
        description: string;
    }>;
    auth: boolean;
}

interface APIRequest {
    method: string;
    url: string;
    headers: Record<string, string>;
    body?: string;
    queryParams: Record<string, string>;
}

interface APIResponse {
    status: number;
    statusText: string;
    headers: Record<string, string>;
    data: any;
    duration: number;
}

interface Environment {
    name: string;
    baseUrl: string;
    apiKey?: string;
    headers: Record<string, string>;
}

const defaultEndpoints: APIEndpoint[] = [
    {
        method: "GET",
        path: "/api/v1/health",
        description: "Health check endpoint",
        parameters: [],
        auth: false
    },
    {
        method: "POST",
        path: "/api/v1/auth/login",
        description: "User login",
        parameters: [
            { name: "email", type: "string", required: true, description: "User email" },
            { name: "password", type: "string", required: true, description: "User password" }
        ],
        auth: false
    },
    {
        method: "GET",
        path: "/api/v1/profile",
        description: "Get user profile",
        parameters: [],
        auth: true
    },
    {
        method: "POST",
        path: "/api/v1/sync/queue",
        description: "Queue offline operation",
        parameters: [
            { name: "operation_type", type: "string", required: true, description: "Type of operation" },
            { name: "table_name", type: "string", required: true, description: "Target table name" },
            { name: "record_id", type: "string", required: false, description: "Record ID" },
            { name: "data", type: "object", required: false, description: "Operation data" }
        ],
        auth: true
    },
    {
        method: "GET",
        path: "/api/v1/sync/status",
        description: "Get sync status",
        parameters: [],
        auth: true
    }
];

const defaultEnvironments: Environment[] = [
    {
        name: "Development",
        baseUrl: "http://localhost:8080",
        headers: {
            "Content-Type": "application/json"
        }
    },
    {
        name: "Staging",
        baseUrl: "https://staging-api.example.com",
        headers: {
            "Content-Type": "application/json"
        }
    },
    {
        name: "Production",
        baseUrl: "https://api.example.com",
        headers: {
            "Content-Type": "application/json"
        }
    }
];

export default function APIExplorer() {
    const [selectedEndpoint, setSelectedEndpoint] = useState<APIEndpoint | null>(null);
    const [request, setRequest] = useState<APIRequest>({
        method: "GET",
        url: "",
        headers: {},
        queryParams: {}
    });
    const [response, setResponse] = useState<APIResponse | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [environments, setEnvironments] = useState<Environment[]>(defaultEnvironments);
    const [selectedEnvironment, setSelectedEnvironment] = useState<Environment>(defaultEnvironments[0]);
    const [authToken, setAuthToken] = useState("");
    const [requestHistory, setRequestHistory] = useState<Array<{ request: APIRequest; response: APIResponse; timestamp: Date }>>([]);
    const [copied, setCopied] = useState(false);

    useEffect(() => {
        if (selectedEndpoint) {
            setRequest(prev => ({
                ...prev,
                method: selectedEndpoint.method,
                url: selectedEnvironment.baseUrl + selectedEndpoint.path,
                headers: {
                    ...selectedEnvironment.headers,
                    ...(authToken && selectedEndpoint.auth ? { "Authorization": `Bearer ${authToken}` } : {})
                }
            }));
        }
    }, [selectedEndpoint, selectedEnvironment, authToken]);

    const handleSendRequest = async () => {
        if (!request.url) {
            toast.error("Please enter a URL");
            return;
        }

        setIsLoading(true);
        const startTime = Date.now();

        try {
            const url = new URL(request.url);
            Object.entries(request.queryParams).forEach(([key, value]) => {
                if (value) url.searchParams.append(key, value);
            });

            const response = await fetch(url.toString(), {
                method: request.method,
                headers: request.headers,
                body: request.body || undefined
            });

            const responseData = await response.json();
            const duration = Date.now() - startTime;

            const apiResponse: APIResponse = {
                status: response.status,
                statusText: response.statusText,
                headers: Object.fromEntries(response.headers.entries()),
                data: responseData,
                duration
            };

            setResponse(apiResponse);
            setRequestHistory(prev => [{
                request: { ...request },
                response: apiResponse,
                timestamp: new Date()
            }, ...prev.slice(0, 49)]); // Keep last 50 requests

            toast.success(`Request completed in ${duration}ms`);
        } catch (error) {
            toast.error(`Request failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
            setResponse(null);
        } finally {
            setIsLoading(false);
        }
    };

    const handleSaveRequest = () => {
        // In a real implementation, this would save to localStorage or a backend
        toast.success("Request saved to collection");
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
        toast.success("Copied to clipboard");
    };

    const getStatusColor = (status: number) => {
        if (status >= 200 && status < 300) return "text-green-600";
        if (status >= 400 && status < 500) return "text-yellow-600";
        if (status >= 500) return "text-red-600";
        return "text-gray-600";
    };

    const getStatusIcon = (status: number) => {
        if (status >= 200 && status < 300) return <CheckCircle className="h-4 w-4 text-green-600" />;
        if (status >= 400 && status < 500) return <AlertCircle className="h-4 w-4 text-yellow-600" />;
        if (status >= 500) return <AlertCircle className="h-4 w-4 text-red-600" />;
        return <Clock className="h-4 w-4 text-gray-600" />;
    };

    return (
        <div className="min-h-screen bg-gray-50 p-6">
            <div className="max-w-7xl mx-auto">
                <div className="mb-8">
                    <h1 className="text-3xl font-bold text-gray-900">API Explorer</h1>
                    <p className="text-gray-600 mt-2">Test and explore your mobile backend API endpoints</p>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    {/* Left Panel - Endpoints and Configuration */}
                    <div className="space-y-6">
                        {/* Environment Selection */}
                        <Card>
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <Globe className="h-5 w-5" />
                                    Environment
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div>
                                    <Label htmlFor="environment">Select Environment</Label>
                                    <Select value={selectedEnvironment.name} onValueChange={(value) => {
                                        const env = environments.find(e => e.name === value);
                                        if (env) setSelectedEnvironment(env);
                                    }}>
                                        <SelectTrigger>
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {environments.map((env) => (
                                                <SelectItem key={env.name} value={env.name}>
                                                    {env.name}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div>
                                    <Label htmlFor="baseUrl">Base URL</Label>
                                    <Input
                                        id="baseUrl"
                                        value={selectedEnvironment.baseUrl}
                                        onChange={(e) => setSelectedEnvironment(prev => ({ ...prev, baseUrl: e.target.value }))}
                                    />
                                </div>
                            </CardContent>
                        </Card>

                        {/* Authentication */}
                        <Card>
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <Key className="h-5 w-5" />
                                    Authentication
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div>
                                    <Label htmlFor="authToken">API Token</Label>
                                    <Input
                                        id="authToken"
                                        type="password"
                                        placeholder="Enter your API token"
                                        value={authToken}
                                        onChange={(e) => setAuthToken(e.target.value)}
                                    />
                                </div>
                            </CardContent>
                        </Card>

                        {/* Available Endpoints */}
                        <Card>
                            <CardHeader>
                                <CardTitle>Available Endpoints</CardTitle>
                                <CardDescription>Select an endpoint to test</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <ScrollArea className="h-64">
                                    <div className="space-y-2">
                                        {defaultEndpoints.map((endpoint, index) => (
                                            <div
                                                key={index}
                                                className={`p-3 rounded-lg border cursor-pointer transition-colors ${selectedEndpoint === endpoint
                                                    ? "border-blue-500 bg-blue-50"
                                                    : "border-gray-200 hover:border-gray-300"
                                                    }`}
                                                onClick={() => setSelectedEndpoint(endpoint)}
                                            >
                                                <div className="flex items-center gap-2 mb-1">
                                                    <Badge variant={endpoint.method === "GET" ? "default" : endpoint.method === "POST" ? "destructive" : "secondary"}>
                                                        {endpoint.method}
                                                    </Badge>
                                                    <span className="text-sm font-mono">{endpoint.path}</span>
                                                    {endpoint.auth && <Key className="h-3 w-3 text-yellow-600" />}
                                                </div>
                                                <p className="text-xs text-gray-600">{endpoint.description}</p>
                                            </div>
                                        ))}
                                    </div>
                                </ScrollArea>
                            </CardContent>
                        </Card>
                    </div>

                    {/* Middle Panel - Request Builder */}
                    <div className="space-y-6">
                        <Card>
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <Send className="h-5 w-5" />
                                    Request Builder
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div className="grid grid-cols-4 gap-2">
                                    <Select value={request.method} onValueChange={(value) => setRequest(prev => ({ ...prev, method: value }))}>
                                        <SelectTrigger>
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="GET">GET</SelectItem>
                                            <SelectItem value="POST">POST</SelectItem>
                                            <SelectItem value="PUT">PUT</SelectItem>
                                            <SelectItem value="DELETE">DELETE</SelectItem>
                                            <SelectItem value="PATCH">PATCH</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <Input
                                        className="col-span-3"
                                        placeholder="Enter URL"
                                        value={request.url}
                                        onChange={(e) => setRequest(prev => ({ ...prev, url: e.target.value }))}
                                    />
                                </div>

                                {selectedEndpoint && selectedEndpoint.parameters.length > 0 && (
                                    <div>
                                        <Label>Parameters</Label>
                                        <div className="space-y-2 mt-2">
                                            {selectedEndpoint.parameters.map((param, index) => (
                                                <div key={index} className="grid grid-cols-3 gap-2">
                                                    <Input
                                                        placeholder={param.name}
                                                        value={request.queryParams[param.name] || ""}
                                                        onChange={(e) => setRequest(prev => ({
                                                            ...prev,
                                                            queryParams: { ...prev.queryParams, [param.name]: e.target.value }
                                                        }))}
                                                    />
                                                    <Select value={param.type} disabled>
                                                        <SelectTrigger>
                                                            <SelectValue />
                                                        </SelectTrigger>
                                                    </Select>
                                                    <div className="flex items-center gap-1">
                                                        <span className="text-xs text-gray-500">{param.required ? "Required" : "Optional"}</span>
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                )}

                                <div>
                                    <Label htmlFor="requestBody">Request Body (JSON)</Label>
                                    <Textarea
                                        id="requestBody"
                                        placeholder="Enter JSON request body"
                                        value={request.body || ""}
                                        onChange={(e) => setRequest(prev => ({ ...prev, body: e.target.value }))}
                                        rows={6}
                                    />
                                </div>

                                <div className="flex gap-2">
                                    <Button onClick={handleSendRequest} disabled={isLoading} className="flex-1">
                                        {isLoading ? (
                                            <>
                                                <Clock className="h-4 w-4 mr-2 animate-spin" />
                                                Sending...
                                            </>
                                        ) : (
                                            <>
                                                <Play className="h-4 w-4 mr-2" />
                                                Send Request
                                            </>
                                        )}
                                    </Button>
                                    <Button variant="outline" onClick={handleSaveRequest}>
                                        <Save className="h-4 w-4" />
                                    </Button>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Response */}
                        {response && (
                            <Card>
                                <CardHeader>
                                    <CardTitle className="flex items-center gap-2">
                                        {getStatusIcon(response.status)}
                                        Response
                                        <Badge variant="outline" className={getStatusColor(response.status)}>
                                            {response.status} {response.statusText}
                                        </Badge>
                                        <Badge variant="secondary">
                                            {response.duration}ms
                                        </Badge>
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <Tabs defaultValue="body" className="w-full">
                                        <TabsList className="grid w-full grid-cols-3">
                                            <TabsTrigger value="body">Body</TabsTrigger>
                                            <TabsTrigger value="headers">Headers</TabsTrigger>
                                            <TabsTrigger value="raw">Raw</TabsTrigger>
                                        </TabsList>
                                        <TabsContent value="body" className="mt-4">
                                            <ScrollArea className="h-64">
                                                <pre className="text-sm bg-gray-100 p-4 rounded-lg overflow-auto">
                                                    {JSON.stringify(response.data, null, 2)}
                                                </pre>
                                            </ScrollArea>
                                        </TabsContent>
                                        <TabsContent value="headers" className="mt-4">
                                            <ScrollArea className="h-64">
                                                <div className="space-y-2">
                                                    {Object.entries(response.headers).map(([key, value]) => (
                                                        <div key={key} className="flex justify-between text-sm">
                                                            <span className="font-mono font-medium">{key}:</span>
                                                            <span className="font-mono text-gray-600">{value}</span>
                                                        </div>
                                                    ))}
                                                </div>
                                            </ScrollArea>
                                        </TabsContent>
                                        <TabsContent value="raw" className="mt-4">
                                            <ScrollArea className="h-64">
                                                <pre className="text-sm bg-gray-100 p-4 rounded-lg overflow-auto">
                                                    {JSON.stringify(response, null, 2)}
                                                </pre>
                                            </ScrollArea>
                                        </TabsContent>
                                    </Tabs>
                                    <div className="mt-4 flex gap-2">
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            onClick={() => copyToClipboard(JSON.stringify(response.data, null, 2))}
                                        >
                                            {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                            Copy Response
                                        </Button>
                                    </div>
                                </CardContent>
                            </Card>
                        )}
                    </div>

                    {/* Right Panel - History and Collections */}
                    <div className="space-y-6">
                        <Card>
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <History className="h-5 w-5" />
                                    Request History
                                </CardTitle>
                                <CardDescription>Recent API requests</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <ScrollArea className="h-64">
                                    <div className="space-y-2">
                                        {requestHistory.map((item, index) => (
                                            <div
                                                key={index}
                                                className="p-3 rounded-lg border cursor-pointer hover:bg-gray-50"
                                                onClick={() => {
                                                    setRequest(item.request);
                                                    setResponse(item.response);
                                                }}
                                            >
                                                <div className="flex items-center gap-2 mb-1">
                                                    <Badge variant={item.request.method === "GET" ? "default" : item.request.method === "POST" ? "destructive" : "secondary"}>
                                                        {item.request.method}
                                                    </Badge>
                                                    <span className="text-sm font-mono truncate">{item.request.url}</span>
                                                </div>
                                                <div className="flex items-center gap-2 text-xs text-gray-500">
                                                    {getStatusIcon(item.response.status)}
                                                    <span className={getStatusColor(item.response.status)}>
                                                        {item.response.status} {item.response.statusText}
                                                    </span>
                                                    <span>{item.response.duration}ms</span>
                                                    <span>{item.timestamp.toLocaleTimeString()}</span>
                                                </div>
                                            </div>
                                        ))}
                                        {requestHistory.length === 0 && (
                                            <p className="text-sm text-gray-500 text-center py-4">No requests yet</p>
                                        )}
                                    </div>
                                </ScrollArea>
                            </CardContent>
                        </Card>

                        {/* Collections */}
                        <Card>
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <Code className="h-5 w-5" />
                                    Collections
                                </CardTitle>
                                <CardDescription>Save and organize requests</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-2">
                                    <div className="p-3 rounded-lg border border-dashed border-gray-300 text-center">
                                        <p className="text-sm text-gray-500">No collections yet</p>
                                        <Button variant="outline" size="sm" className="mt-2">
                                            Create Collection
                                        </Button>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        </div>
    );
}
