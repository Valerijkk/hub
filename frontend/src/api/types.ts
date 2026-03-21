/** Paginated response from the API. */
export interface PaginatedResponse<T> {
  items: T[];
  next_cursor?: string;
  has_more?: boolean;
  total: number;
  offset?: number;
  limit?: number;
}
