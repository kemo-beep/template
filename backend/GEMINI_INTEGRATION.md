# Google Gemini AI Integration

This document describes the Google Gemini AI integration in the mobile backend API.

## Overview

The Gemini AI integration provides text generation capabilities using Google's Gemini AI models. It includes conversation management, context-aware text generation, and comprehensive API endpoints.

## Features

- **Text Generation**: Generate text using various Gemini AI models
- **Conversation Management**: Create, manage, and maintain conversation history
- **Context-Aware Generation**: Generate text with conversation context
- **Multiple Models**: Support for different Gemini models (gemini-1.5-flash, gemini-1.5-pro, etc.)
- **Caching**: Response caching for improved performance
- **Rate Limiting**: Built-in rate limiting for API protection
- **Monitoring**: Comprehensive logging and metrics

## Configuration

### Environment Variables

Add the following environment variables to your `.env` file:

```bash
# Google Gemini AI Configuration
GEMINI_API_KEY=your_gemini_api_key
GEMINI_MODEL=gemini-1.5-flash
GEMINI_MAX_TOKENS=8192
GEMINI_TEMPERATURE=0.7
GEMINI_TOP_P=0.8
GEMINI_TOP_K=40
```

### Getting API Key

1. Visit [Google AI Studio](https://aistudio.google.com/)
2. Sign in with your Google account
3. Create a new API key
4. Copy the API key to your environment variables

## API Endpoints

### Public Endpoints

- `GET /api/v1/gemini/health` - Health check for Gemini service
- `GET /api/v1/gemini/models` - Get available models

### Protected Endpoints (Require Authentication)

#### Text Generation
- `POST /api/v1/gemini/generate` - Generate text
- `POST /api/v1/gemini/conversations/{conversation_id}/generate` - Generate text with context

#### Conversation Management
- `POST /api/v1/gemini/conversations` - Create conversation
- `GET /api/v1/gemini/conversations` - List conversations
- `GET /api/v1/gemini/conversations/{conversation_id}` - Get conversation
- `DELETE /api/v1/gemini/conversations/{conversation_id}` - Delete conversation

#### Message Management
- `POST /api/v1/gemini/conversations/{conversation_id}/messages` - Add message

#### Service Management
- `GET /api/v1/gemini/stats` - Get service statistics

## Usage Examples

### Generate Text

```bash
curl -X POST http://localhost:8080/api/v1/gemini/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "prompt": "Explain quantum computing in simple terms",
    "model": "gemini-1.5-flash",
    "max_tokens": 1000,
    "temperature": 0.7
  }'
```

### Create Conversation

```bash
curl -X POST http://localhost:8080/api/v1/gemini/conversations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Quantum Computing Discussion"
  }'
```

### Generate Text with Context

```bash
curl -X POST http://localhost:8080/api/v1/gemini/conversations/{conversation_id}/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "prompt": "What are the practical applications?",
    "context": "We are discussing quantum computing"
  }'
```

## Models

The integration supports the following Gemini models:

- `gemini-1.5-flash` (default) - Fast and efficient
- `gemini-1.5-pro` - More capable and accurate
- `gemini-1.0-pro` - Legacy model

## Parameters

### Generation Parameters

- `prompt` (required): The text prompt for generation
- `model` (optional): Gemini model to use
- `max_tokens` (optional): Maximum tokens to generate
- `temperature` (optional): Randomness in generation (0.0-1.0)
- `top_p` (optional): Nucleus sampling parameter
- `top_k` (optional): Top-k sampling parameter
- `context` (optional): Additional context for generation
- `metadata` (optional): Custom metadata

### Conversation Parameters

- `title` (required): Conversation title
- `role` (required): Message role ("user" or "assistant")
- `content` (required): Message content

## Database Schema

### GeminiConversation

```sql
CREATE TABLE gemini_conversations (
    id SERIAL PRIMARY KEY,
    id VARCHAR(255) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    messages JSONB,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Rate Limiting

The Gemini endpoints are protected by rate limiting:

- Text generation: Standard API rate limit
- Conversation management: Standard API rate limit
- Health check: No rate limiting

## Monitoring

The integration includes comprehensive logging and metrics:

- Request/response logging
- Error tracking
- Performance metrics
- Cache statistics
- Service health monitoring

## Security

- API key is stored securely in environment variables
- All endpoints require authentication (except health check and models)
- Rate limiting prevents abuse
- Input validation and sanitization

## Troubleshooting

### Common Issues

1. **API Key Error**: Ensure `GEMINI_API_KEY` is set correctly
2. **Model Not Found**: Check if the model name is correct
3. **Rate Limit Exceeded**: Wait and retry, or check rate limit configuration
4. **Database Connection**: Ensure database is running and accessible

### Debug Mode

Enable debug logging by setting `LOG_LEVEL=debug` in your environment variables.

## Future Enhancements

- Streaming responses
- Image generation support
- Multi-modal capabilities
- Advanced conversation features
- Custom model fine-tuning
