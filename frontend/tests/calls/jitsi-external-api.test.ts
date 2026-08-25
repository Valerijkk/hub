import { afterEach, describe, expect, it, vi } from "vitest";

import { getJitsiExternalAPIScriptURL } from "@/modules/calls/jitsi-external-api";

const maxLengthHost = [
  "meet",
  "a".repeat(63),
  "b".repeat(63),
  "c".repeat(63),
  "d".repeat(56),
].join(".");

const tooLongHost = `${maxLengthHost}e`;
const invalidOriginMessage = "Invalid Jitsi external API origin";

const invalidOrigins = [
  ["non-HTTPS scheme", "http://meet.hub.example.com"],
  ["leading whitespace", " https://meet.hub.example.com"],
  ["trailing whitespace", "https://meet.hub.example.com "],
  ["root slash", "https://meet.hub.example.com/"],
  ["path", "https://meet.hub.example.com/tenant"],
  ["query", "https://meet.hub.example.com?tenant=one"],
  ["fragment", "https://meet.hub.example.com#tenant"],
  ["explicit default port", "https://meet.hub.example.com:443"],
  ["non-default port", "https://meet.hub.example.com:8443"],
  ["userinfo", "https://user:password@meet.hub.example.com"],
  ["uppercase scheme", "HTTPS://meet.hub.example.com"],
  ["uppercase hostname", "https://MEET.hub.example.com"],
  ["normalized unicode hostname", "https://meet.h\u00fcb.example.com"],
  ["IPv4 hostname", "https://127.0.0.1"],
  ["IPv6 hostname", "https://[::1]"],
  ["localhost", "https://localhost"],
  ["bare meet hostname", "https://meet"],
  ["bare meet hostname with root dot", "https://meet."],
  ["canonical hostname with root dot", "https://meet.hub.example.com."],
  ["non-meet prefix", "https://video.hub.example.com"],
  ["vendor origin", "https://meet.jit.si"],
  ["vendor tenant subdomain", "https://tenant.meet.jit.si"],
  ["empty hostname label", "https://meet..hub.example.com"],
  ["underscore hostname label", "https://meet.qa_team.example.com"],
  ["leading hyphen in hostname label", "https://meet.-qa.example.com"],
  ["trailing hyphen in hostname label", "https://meet.qa-.example.com"],
  ["hostname label longer than 63 octets", `https://meet.${"a".repeat(64)}.example.com`],
  ["hostname longer than 253 octets", `https://${tooLongHost}`],
] as const;

describe("getJitsiExternalAPIScriptURL", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("accepts a canonical HTTPS meet origin", () => {
    expect(getJitsiExternalAPIScriptURL("https://meet.hub.example.com")).toBe(
      "https://meet.hub.example.com/external_api.js",
    );
  });

  it("allows a self-hosted meet origin under jit.si without broad suffix blocking", () => {
    const origin = "https://meet.hub.jit.si";

    expect(getJitsiExternalAPIScriptURL(origin)).toBe(`${origin}/external_api.js`);
  });
  it("accepts a hostname at the 253-octet boundary", () => {
    const origin = `https://${maxLengthHost}`;

    expect(getJitsiExternalAPIScriptURL(origin)).toBe(`${origin}/external_api.js`);
  });

  it.each(invalidOrigins)("rejects %s with the fixed public error", (_caseName, origin) => {
    let thrown: unknown;

    try {
      getJitsiExternalAPIScriptURL(origin);
    } catch (error) {
      thrown = error;
    }

    expect(thrown).toBeInstanceOf(Error);
    expect((thrown as Error).message).toBe(invalidOriginMessage);
    expect((thrown as Error).message).not.toContain(origin);
  });

  it("uses one generic validation error that never echoes rejected origins", () => {
    const rejectedOrigins = [
      "https://user:password@meet.hub.example.com",
      `https://${tooLongHost}`,
    ];

    const messages = rejectedOrigins.map((origin) => {
      let thrown: unknown;

      try {
        getJitsiExternalAPIScriptURL(origin);
      } catch (error) {
        thrown = error;
      }

      expect(thrown).toBeInstanceOf(Error);
      expect((thrown as Error).message).toBe(invalidOriginMessage);
      expect((thrown as Error).message).not.toContain(origin);
      return (thrown as Error).message;
    });

    expect(new Set(messages)).toEqual(new Set([invalidOriginMessage]));
  });

  it("is pure for accepted and rejected origins and ignores ambient browser location", () => {
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

    expect(getJitsiExternalAPIScriptURL("https://meet.hub.example.com")).toBe(
      "https://meet.hub.example.com/external_api.js",
    );
    expect(() => getJitsiExternalAPIScriptURL("http://meet.hub.example.com")).toThrowError(
      invalidOriginMessage,
    );
    expect(fetchSpy).not.toHaveBeenCalled();
    expect(createElementSpy).not.toHaveBeenCalled();
    expect(appendChildSpy).not.toHaveBeenCalled();
  });
});