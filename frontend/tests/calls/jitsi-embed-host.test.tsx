import { cleanup, render } from "@testing-library/react";
import { createRef } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { JitsiEmbedHost } from "@/modules/calls/jitsi-embed-host";

const validCredential = {
  provider: "jitsi",
  roomName: "team-call-42",
  serviceURL: "https://meet.hub.example.com",
  token: "token-must-stay-out-of-dom",
};

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe("JitsiEmbedHost", () => {
  it("renders one empty inert host and forwards its ref for a valid credential", () => {
    const ref = createRef<HTMLDivElement>();
    const { container } = render(
      <JitsiEmbedHost ref={ref} credential={validCredential} />,
    );

    const hosts = container.querySelectorAll("[data-jitsi-embed-host]");
    expect(hosts).toHaveLength(1);
    expect(ref.current).toBe(hosts[0]);
    expect(hosts[0]).toBeEmptyDOMElement();
    expect(hosts[0].getAttributeNames()).toEqual(["data-jitsi-embed-host"]);
    expect(hosts[0]).toHaveAttribute("data-jitsi-embed-host", "");
    expect(container.querySelector("script, iframe")).toBeNull();
    expect(container.innerHTML).not.toContain(validCredential.roomName);
    expect(container.innerHTML).not.toContain(validCredential.serviceURL);
    expect(container.innerHTML).not.toContain(validCredential.token);
  });

  it.each([
    ["missing credential", null],
    ["wrong provider", { ...validCredential, provider: "livekit" }],
    ["blank room", { ...validCredential, roomName: "" }],
    ["padded room", { ...validCredential, roomName: " team-call-42 " }],
    ["blank token", { ...validCredential, token: " " }],
    ["insecure service URL", { ...validCredential, serviceURL: "http://meet.hub.example.com" }],
    ["public Jitsi service", { ...validCredential, serviceURL: "https://meet.jit.si" }],
  ])("fails closed for %s", (_caseName, credential) => {
    const ref = createRef<HTMLDivElement>();
    const { container } = render(
      <JitsiEmbedHost ref={ref} credential={credential} />,
    );

    expect(container).toBeEmptyDOMElement();
    expect(ref.current).toBeNull();
  });

  it("removes the host and clears the ref when credentials become invalid", () => {
    const ref = createRef<HTMLDivElement>();
    const { container, rerender } = render(
      <JitsiEmbedHost ref={ref} credential={validCredential} />,
    );
    expect(ref.current).not.toBeNull();

    rerender(
      <JitsiEmbedHost
        ref={ref}
        credential={{ ...validCredential, token: " " }}
      />,
    );

    expect(container).toBeEmptyDOMElement();
    expect(ref.current).toBeNull();
  });

  it("does not activate ambient capabilities or leak sensitive values", () => {
    const fetchSpy = vi.fn();
    const externalAPISpy = vi.fn();
    const consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    const consoleLogSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    const storageGetSpy = vi.spyOn(Storage.prototype, "getItem");
    const storageSetSpy = vi.spyOn(Storage.prototype, "setItem");
    const storageRemoveSpy = vi.spyOn(Storage.prototype, "removeItem");
    const storageClearSpy = vi.spyOn(Storage.prototype, "clear");
    vi.stubGlobal("fetch", fetchSpy);
    vi.stubGlobal("JitsiMeetExternalAPI", externalAPISpy);

    const { container } = render(
      <JitsiEmbedHost credential={validCredential} />,
    );

    expect(fetchSpy).not.toHaveBeenCalled();
    expect(externalAPISpy).not.toHaveBeenCalled();
    expect(storageGetSpy).not.toHaveBeenCalled();
    expect(storageSetSpy).not.toHaveBeenCalled();
    expect(storageRemoveSpy).not.toHaveBeenCalled();
    expect(storageClearSpy).not.toHaveBeenCalled();
    expect(consoleErrorSpy).not.toHaveBeenCalled();
    expect(consoleLogSpy).not.toHaveBeenCalled();
    expect(container.querySelector("script, iframe")).toBeNull();
    expect(container.innerHTML).not.toContain(validCredential.roomName);
    expect(container.innerHTML).not.toContain(validCredential.serviceURL);
    expect(container.innerHTML).not.toContain(validCredential.token);
  });
});
