type RequestOptions = {
  allowUnauthorized?: boolean;
};

export async function apiRequest<T = unknown>(
  url: string,
  init: RequestInit = {},
  options: RequestOptions = {},
): Promise<{ response: Response; data: T }> {
  const { allowUnauthorized = false } = options;
  const response = await fetch(url, {
    credentials: "same-origin",
    ...init,
  });

  const raw = await response.text();
  let data: T;

  if (!raw) {
    data = {} as T;
  } else {
    try {
      data = JSON.parse(raw) as T;
    } catch {
      data = ({ raw } as unknown) as T;
    }
  }

  if (!allowUnauthorized && response.status === 401) {
    throw new Error("Session expired. Please log in again.");
  }

  return { response, data };
}
