import { afterEach, describe, expect, it, vi } from "vitest";

import { buildJitsiEmbedContract } from "@/modules/calls/jitsi-embed-contract";

const invalidContractMessage = "Invalid Jitsi embed contract";

const validInput = {
  provider: "jitsi",
  roomName: "team-call-42",
  serviceURL: "https://meet.hub.example.com",
  token: "header.payload.signature",
};

const invalidInputs = [
  ["non-Jitsi provider", { ...validInput, provider: "livekit" }],
  ["blank room", { ...validInput, roomName: "" }],
  ["whitespace-only room", { ...validInput, roomName: "   " }],
  ["leading-padded room", { ...validInput, roomName: " team-call-42" }],
  ["trailing-padded room", { ...validInput, roomName: "team-call-42 " }],
  ["blank token", { ...validInput, token: "" }],
  ["whitespace-only token", { ...validInput, token: "   " }],
  ["leading-padded token", { ...validInput, token: " header.payload.signature" }],
  ["trailing-padded token", { ...validInput, token: "header.payload.signature " }],
  ["non-HTTPS origin", { ...validInput, serviceURL: "http://meet.hub.example.com" }],
  ["vendor origin", { ...validInput, serviceURL: "https://meet.jit.si" }],
] as const;

function expectGenericContractError(action: () => unknown, input: (typeof invalidInputs)[number][1]) {
  let thrown: unknown;

  try {
    action();
  } catch (error) {
    thrown = error;
  }

  expect(thrown).toBeInstanceOf(Error);
  expect((thrown as Error).message).toBe(invalidContractMessage);
  for (const value of [input.provider, input.serviceURL, input.roomName, input.token]) {
    if (value.length > 0) {
      expect((thrown as Error).message).not.toContain(value);
    }
  }
}

describe("buildJitsiEmbedContract", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("builds the canonical self-hosted Jitsi embed contract", () => {
    expect(buildJitsiEmbedContract(validInput)).toEqual({
      externalApiScriptURL: "https://meet.hub.example.com/external_api.js",
      domain: "meet.hub.example.com",
      roomName: "team-call-42",
      jwt: "header.payload.signature",
    });
  });

  it.each(invalidInputs)("rejects %s with one nonleaking generic error", (_caseName, input) => {
    expectGenericContractError(() => buildJitsiEmbedContract(input), input);
  });

  it("does not read ambient browser globals or create network and DOM effects", () => {
    const fetchSpy = vi.fn();
    const createElementSpy = vi.fn();
    const appendChildSpy = vi.fn();
    const forbiddenLocation = {
      get href() {
        throw new Error("ambient location href must not be read");
      },
      get origin() {
        throw new Error("ambient location origin must not be read");
      },
    };

    vi.stubGlobal("fetch", fetchSpy);
    vi.stubGlobal("document", {
      body: { appendChild: appendChildSpy },
      createElement: createElementSpy,
      documentElement: { appendChild: appendChildSpy },
      head: { appendChild: appendChildSpy },
    });
    vi.stubGlobal("location", forbiddenLocation);
    vi.stubGlobal("window", { location: forbiddenLocation });

    expect(buildJitsiEmbedContract(validInput)).toEqual({
      externalApiScriptURL: "https://meet.hub.example.com/external_api.js",
      domain: "meet.hub.example.com",
      roomName: "team-call-42",
      jwt: "header.payload.signature",
    });
    expectGenericContractError(
      () => buildJitsiEmbedContract({ ...validInput, serviceURL: "https://meet.jit.si" }),
      { ...validInput, serviceURL: "https://meet.jit.si" },
    );
    expect(fetchSpy).not.toHaveBeenCalled();
    expect(createElementSpy).not.toHaveBeenCalled();
    expect(appendChildSpy).not.toHaveBeenCalled();
  });

  it("never logs the supplied Jitsi token", () => {
    const token = "token-must-not-be-logged";
    const consoleSpies = ["debug", "error", "info", "log", "warn"].map((method) =>
      vi.spyOn(console, method).mockImplementation(() => undefined),
    );

    buildJitsiEmbedContract({ ...validInput, token });

    for (const consoleSpy of consoleSpies) {
      expect(consoleSpy).not.toHaveBeenCalled();
    }
  });
});
