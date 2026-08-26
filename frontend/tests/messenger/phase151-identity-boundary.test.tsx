import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { getUnreadCounts, markChannelRead } from "@/api/messenger";
import { unreadQueryKey, useMarkRead, useUnread } from "@/hooks/use-unread";
import { useAuthStore } from "@/stores/auth-store";
import { emptyUnreadSnapshot, useMessengerStore } from "@/stores/messenger-store";

vi.mock("@/api/messenger", () => ({
  getUnreadCounts: vi.fn(),
  markChannelRead: vi.fn(),
}));

type UnreadSnapshot = ReturnType<typeof emptyUnreadSnapshot>;

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });

  return { promise, reject, resolve };
}

const userA = {
  id: "user-a",
  email: "user-a@example.test",
  timezone: "UTC",
  language: "en",
  system_role: "member",
  onboarding_completed: true,
};

const userB = {
  id: "user-b",
  email: "user-b@example.test",
  timezone: "UTC",
  language: "en",
  system_role: "member",
  onboarding_completed: true,
};

const queryClients: QueryClient[] = [];

function createQueryClient() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  queryClients.push(queryClient);
  return queryClient;
}

function createWrapper(queryClient: QueryClient) {
  return function QueryWrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

describe("Phase 151 identity boundary for unread state", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ user: userA });
    useMessengerStore.getState().reset();
    useMessengerStore.getState().setUnreadSnapshot(emptyUnreadSnapshot());
  });

  afterEach(() => {
    for (const queryClient of queryClients.splice(0)) {
      queryClient.clear();
    }
  });

  it("never commits a deferred old-user unread snapshot after identity switches", async () => {
    const queryClient = createQueryClient();
    const unreadForA = deferred<UnreadSnapshot>();
    const unreadForB = deferred<UnreadSnapshot>();
    const snapshotA: UnreadSnapshot = { channels: { "channel-a": 3 }, dms: {}, mentions: {}, threads: {}, any_unread: true };
    const snapshotB: UnreadSnapshot = { channels: { "channel-b": 5 }, dms: {}, mentions: {}, threads: {}, any_unread: true };

    vi.mocked(getUnreadCounts)
      .mockReturnValueOnce(unreadForA.promise)
      .mockReturnValueOnce(unreadForB.promise);

    const { rerender } = renderHook(() => useUnread(), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(getUnreadCounts).toHaveBeenCalledTimes(1));

    act(() => {
      useAuthStore.setState({ user: userB });
    });
    rerender();

    await waitFor(() => expect(getUnreadCounts).toHaveBeenCalledTimes(2));

    await act(async () => {
      unreadForA.resolve(snapshotA);
      await unreadForA.promise;
    });

    expect(useMessengerStore.getState().unreadSnapshot).not.toEqual(snapshotA);
    expect(queryClient.getQueryData(unreadQueryKey(userA.id))).toBe(snapshotA);

    await act(async () => {
      unreadForB.resolve(snapshotB);
      await unreadForB.promise;
    });

    await waitFor(() => {
      expect(useMessengerStore.getState().unreadSnapshot).toEqual(snapshotB);
    });
    expect(queryClient.getQueryData(unreadQueryKey(userB.id))).toBe(snapshotB);
  });

  it("does not invalidate either identity after an in-flight mark-read completes across a switch", async () => {
    const queryClient = createQueryClient();
    const markRead = deferred<void>();
    const snapshotB: UnreadSnapshot = { channels: { "channel-b": 5 }, dms: {}, mentions: {}, threads: {}, any_unread: true };
    const invalidateQueries = vi.spyOn(queryClient, "invalidateQueries");

    vi.mocked(markChannelRead).mockReturnValueOnce(markRead.promise);

    const { result, rerender } = renderHook(() => useMarkRead(), {
      wrapper: createWrapper(queryClient),
    });

    let markPromise!: Promise<void>;
    act(() => {
      markPromise = result.current("channel-a");
    });

    await waitFor(() => expect(markChannelRead).toHaveBeenCalledWith("channel-a"));

    act(() => {
      useAuthStore.setState({ user: userB });
      useMessengerStore.getState().setUnreadSnapshot(snapshotB);
    });
    rerender();

    await act(async () => {
      markRead.resolve();
      await markPromise;
    });

    expect(useMessengerStore.getState().unreadSnapshot).toEqual(snapshotB);
    expect(invalidateQueries).not.toHaveBeenCalledWith({
      exact: true,
      queryKey: unreadQueryKey(userA.id),
    });
    expect(invalidateQueries).not.toHaveBeenCalledWith({
      exact: true,
      queryKey: unreadQueryKey(userB.id),
    });
  });

  it("disables unread fetching and fails closed to the empty snapshot without an identity", async () => {
    const queryClient = createQueryClient();
    const staleSnapshot: UnreadSnapshot = { channels: { stale: 9 }, dms: {}, mentions: {}, threads: {}, any_unread: true };

    act(() => {
      useAuthStore.setState({ user: null });
      useMessengerStore.getState().setUnreadSnapshot(staleSnapshot);
    });

    renderHook(() => useUnread(), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => {
      expect(useMessengerStore.getState().unreadSnapshot).toEqual(emptyUnreadSnapshot());
    });
    expect(getUnreadCounts).not.toHaveBeenCalled();
  });
});
