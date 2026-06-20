import { describe, expect, it } from "vitest";
import enCommon from "./locales/en/common.json";
import ruCommon from "./locales/ru/common.json";

const validationKeys = [
	"enterEmail",
	"enterCode",
	"enterComment",
	"enterQuestion",
	"pollMinOptions",
	"enterLabelName",
	"enterGroupName",
	"nameCannotBeEmpty",
	"fillNameAndKey",
	"enterNameForKeyGen",
	"avatarMaxSize",
	"avatarImagesOnly",
	"enterMinutes",
	"enterStatusName",
	"selectBothStatuses",
	"statusesMustDiffer",
	"pushPermissionDenied",
	"pushSubscribeFailed",
	"objectNotFound",
] as const;

describe("common validation i18n keys", () => {
	it("has all validation keys in en and ru", () => {
		for (const key of validationKeys) {
			expect(enCommon.validation).toHaveProperty(key);
			expect(ruCommon.validation).toHaveProperty(key);
			expect(enCommon.validation[key as keyof typeof enCommon.validation]).toBeTruthy();
			expect(ruCommon.validation[key as keyof typeof ruCommon.validation]).toBeTruthy();
		}
	});
});
