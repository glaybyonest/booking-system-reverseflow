export type ApiList<T> = {
  items: T[];
  total?: number;
};

export type ApiErrorPayload = {
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
};
