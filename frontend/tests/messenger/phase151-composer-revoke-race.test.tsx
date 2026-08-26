import { act, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { MessageComposer } from "@/components/messenger/message-composer";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

import { renderWithProviders } from "../support/render";

const mocks = vi.hoisted(() => ({
  sendMessage: vi.fn(),
  uploadFile: vi.fn(),
  toastDismiss: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock("@/api/messenger", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/api/messenger")>();

  return {
    ...actual,
    sendMessage: mocks.sendMessage,
    sendTypingIndicator: vi.fn(),
  };
});

vi.mock("@/api/files", () => ({
  uploadFile: mocks.uploadFile,
}));

vi.mock("sonner", () => ({
  toast: {
    dismiss: mocks.toastDismiss,
    error: mocks.toastError,
  },
}));

describe("MessageComposer revoke race barrier", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useMessengerStore.getState().reset();
    useAuthStore.setState({
      user: {
        id: "composer-revoke-user",
        email: "composer-revoke-user@example.test",
        timezone: "UTC",
        language: "en",
        system_role: "member",
        onboarding_completed: true,
      },
      isAuthenticated: true,
    });
    mocks.toastError.mockReturnValue("retry-toast");
  });

  afterEach(() => {
    useAuthStore.setState({ user: null, isAuthenticated: false });
  });

  it("does not retry a rejected send after its channel has been revoked", async () => {
    mocks.sendMessage.mockRejectedValueOnce(new Error("network unavailable"));

    renderWithProviders(<MessageComposer canSend channelId="channel-revoked" />);

    const textbox = screen.getByRole("textbox");
    fireEvent.change(textbox, { target: { value: "Message before revoke" } });
    fireEvent.keyDown(textbox, { key: "Enter", code: "Enter" });

    await waitFor(() => expect(mocks.sendMessage).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(mocks.toastError).toHaveBeenCalledTimes(1));

    const options = mocks.toastError.mock.calls[0]?.[1] as {
      action: { onClick: () => void };
    };

    useMessengerStore.getState().revokeChannel("channel-revoked");
    options.action.onClick();

    expect(mocks.sendMessage).toHaveBeenCalledTimes(1);
  });

  it("discards an upload that settles after its channel has been revoked", async () => {
    let resolveUpload: (fileID: string) => void = () => undefined;
    mocks.uploadFile.mockImplementationOnce(
      () =>
        new Promise<string>((resolve) => {
          resolveUpload = resolve;
        }),
    );

    renderWithProviders(<MessageComposer canSend channelId="channel-revoked" />);

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    expect(fileInput).toBeTruthy();
    fireEvent.change(fileInput, {
      target: {
        files: [new File(["x"], "x.txt", { type: "text/plain" })],
      },
    });

    await waitFor(() => expect(mocks.uploadFile).toHaveBeenCalledTimes(1));
    useMessengerStore.getState().revokeChannel("channel-revoked");

    await act(async () => {
      resolveUpload("uploaded-file-id");
      await Promise.resolve();
    });

    expect(screen.queryByText("x.txt")).not.toBeInTheDocument();

    const textbox = screen.getByRole("textbox");
    fireEvent.change(textbox, { target: { value: "Message after revoke" } });
    fireEvent.keyDown(textbox, { key: "Enter", code: "Enter" });

    expect(mocks.sendMessage).not.toHaveBeenCalled();
  });
});
