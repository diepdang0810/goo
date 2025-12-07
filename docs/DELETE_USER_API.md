# DELETE User API Documentation

## Endpoint
`DELETE /users/{id}`

## Description
Deletes a user from the system by their unique identifier. This operation will:
1. Remove the user record from the PostgreSQL database
2. Remove the user from Redis cache
3. Return success message if deletion is successful

## Request

### Path Parameters
| Parameter | Type   | Required | Description                    |
|-----------|--------|----------|--------------------------------|
| id        | int64  | Yes      | The unique identifier of user  |

### Headers
No special headers required.

### Request Example
```bash
curl -X DELETE http://localhost:3002/users/1
```

## Response

### Success Response (200 OK)
```json
{
  "code": 200,
  "data": {
    "message": "User deleted successfully"
  }
}
```

### Error Responses

#### Invalid ID (400 Bad Request)
```json
{
  "code": 400,
  "error": "Invalid ID"
}
```

#### User Not Found (404 Not Found)
```json
{
  "code": 1002,
  "error": "user not found"
}
```

#### Internal Server Error (500 Internal Server Error)
```json
{
  "code": 500,
  "error": "internal server error"
}
```

## Error Codes
| Code | HTTP Status | Description                    |
|------|-------------|--------------------------------|
| 200  | 200         | Success                        |
| 400  | 400         | Invalid ID format              |
| 1002 | 404         | User not found                 |
| 500  | 500         | Internal server error          |

## Testing Examples

### Using cURL
```bash
# Delete user with ID 1
curl -X DELETE http://localhost:3002/users/1

# Expected Success Response:
# {"code":200,"data":{"message":"User deleted successfully"}}

# Delete non-existent user
curl -X DELETE http://localhost:3002/users/999

# Expected Error Response:
# {"code":1002,"error":"user not found"}

# Delete with invalid ID
curl -X DELETE http://localhost:3002/users/invalid

# Expected Error Response:
# {"code":400,"error":"Invalid ID"}
```

### Using HTTPie
```bash
# Delete user
http DELETE http://localhost:3002/users/1
```

### Using Postman
1. Method: `DELETE`
2. URL: `http://localhost:3002/users/1`
3. Click Send

## Implementation Notes

1. **Database Operation**: Uses `DELETE FROM users WHERE id = $1` SQL query
2. **Cache Invalidation**: Automatically removes `user:{id}` key from Redis
3. **Error Handling**:
   - If cache deletion fails, it will be logged but won't fail the entire operation
   - If database deletion fails, the operation returns an error immediately
4. **Transaction**: No transaction needed as it's a single DELETE operation
5. **Idempotency**: Deleting an already-deleted user returns 404 error

## Related Endpoints

- `POST /users` - Create a new user
- `GET /users/{id}` - Get user by ID
- `GET /users` - List all users

## Import to Apidog

To import this API specification into Apidog:

1. Open Apidog application
2. Navigate to your project
3. Click on "Import" button
4. Select "OpenAPI/Swagger" format
5. Upload the file: `docs/openapi-delete-user.yaml`
6. Click "Import"

The DELETE user endpoint will be added to your Apidog project with:
- Complete request/response documentation
- Example requests and responses
- Error scenarios
- Parameter validation rules
