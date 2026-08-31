//go:build qa && !release

package qamailcalendar

// inboundRequestFixtureExternalClosureManifestReplica is a literal replica of
// the frozen external-closure policy. The committed machine manifest remains
// the normative policy authority; this value exists only for C4b/C4k static
// verification and must be updated in the same reviewed change as that asset.
//
// The raw literal is deliberately LF-terminated. Keep its bytes, order, and
// six-field grammar exact; do not derive it from schemas, catalogs, reflection,
// files, maps, or slices at runtime.
const inboundRequestFixtureExternalClosureManifestReplica = `Organization|calendar_availability_reservations|organization_id|forbidden|none|0
Organization|calendar_events|organization_id|allowlisted|calendarProof.calendarEventID|1
Organization|calendar_inbound_object_states|organization_id|allowlisted|calendarProof.inboundObjectStateID|1
Organization|calendar_invite_deliveries|organization_id|forbidden|none|0
Organization|calendar_mutation_receipts|organization_id|allowlisted|calendarProof.mutationReceiptID|1
Organization|calendar_reply_deliveries|organization_id|allowlisted|calendarProof.replyDeliveryID|1
Organization|mail_accounts|organization_id|allowlisted|rootProof.mailAccountID|1
Organization|mail_domains|organization_id|allowlisted|rootProof.mailDomainID|1
Organization|mail_mutation_receipts|organization_id|forbidden|none|0
Organization|mail_objects|organization_id|forbidden|none|0
Organization|mail_send_rate_buckets|organization_id|forbidden|none|0
Organization|mailbox_policies|organization_id|forbidden|none|0
Organization|projects|organization_projects|forbidden|none|0
Organization|users|organization_users|allowlisted|rootProof.ownerUserID|1
User|activity_logs|user_activity_logs|forbidden|none|0
User|ai_providers|user_ai_providers|forbidden|none|0
User|ai_usage_dailies|user_ai_usage_daily|forbidden|none|0
User|ai_user_preferences|user_ai_preference|forbidden|none|0
User|api_keys|user_api_keys|forbidden|none|0
User|automation_rules|created_by_id|forbidden|none|0
User|bookmarks|user_id|forbidden|none|0
User|calendar_attendees|user_id|allowlisted|calendarProof.rsvpAttendeeID|1
User|calendar_availability_reservations|user_id|forbidden|none|0
User|calendar_events|calendar_owner_id|allowlisted|calendarProof.calendarEventID|1
User|calendar_events|organizer_id|forbidden|none|0
User|calendar_inbound_object_states|calendar_owner_id|allowlisted|calendarProof.inboundObjectStateID|1
User|calendar_mutation_receipts|actor_user_id|allowlisted|calendarProof.mutationReceiptID|1
User|call_participants|user_id|forbidden|none|0
User|calls|created_by|forbidden|none|0
User|channel_invites|creator_id|forbidden|none|0
User|channel_members|user_id|forbidden|none|0
User|channels|created_by|forbidden|none|0
User|coffee_enrollments|user_coffee_enrollment|forbidden|none|0
User|coffee_pairs|user_coffee_pairs_a|forbidden|none|0
User|coffee_pairs|user_coffee_pairs_b|forbidden|none|0
User|coffee_rounds|user_coffee_rounds_triggered|forbidden|none|0
User|comments|user_comments|forbidden|none|0
User|crm_contacts|user_crm_contact_link|forbidden|none|0
User|crm_contacts|user_crm_contacts_created|forbidden|none|0
User|crm_deals|user_crm_deals_owned|forbidden|none|0
User|dashboard_layouts|user_id|forbidden|none|0
User|file_metadata|uploader_id|forbidden|none|0
User|hub_bases|user_hub_bases|forbidden|none|0
User|hub_records|user_hub_records|forbidden|none|0
User|invites|user_invites|forbidden|none|0
User|invoices|user_invoices_created|forbidden|none|0
User|issue_message_links|created_by_id|forbidden|none|0
User|issues|issue_assignee|forbidden|none|0
User|issues|issue_reporter|forbidden|none|0
User|journal_entries|user_journal_entries|forbidden|none|0
User|kudos|user_kudos_received|forbidden|none|0
User|kudos|user_kudos_sent|forbidden|none|0
User|leave_requests|user_leave_requests|forbidden|none|0
User|leave_requests|user_leave_reviews|forbidden|none|0
User|mail_accounts|user_id|allowlisted|rootProof.mailAccountID|1
User|mail_objects|uploader_user_id|forbidden|none|0
User|mail_send_rate_buckets|user_id|forbidden|none|0
User|marketplace_listings|user_marketplace_listings|forbidden|none|0
User|marketplace_listings|user_marketplace_reviews|forbidden|none|0
User|news_poll_votes|user_id|forbidden|none|0
User|news_posts|user_news_posts|forbidden|none|0
User|notification_preferences|user_id|forbidden|none|0
User|notifications|user_id|forbidden|none|0
User|poll_votes|user_id|forbidden|none|0
User|polls|creator_id|forbidden|none|0
User|project_members|user_project_members|forbidden|none|0
User|project_templates|user_project_templates|forbidden|none|0
User|push_subscriptions|user_id|forbidden|none|0
User|reactions|user_reactions|forbidden|none|0
User|recent_contacts|user_id|forbidden|none|0
User|recordings|started_by|forbidden|none|0
User|saved_filters|owner_id|forbidden|none|0
User|scheduled_calls|created_by|forbidden|none|0
User|sessions|user_sessions|forbidden|none|0
User|team_members|user_team_members|forbidden|none|0
User|test_case_versions|user_test_case_versions|forbidden|none|0
User|test_cases|user_created_test_cases|forbidden|none|0
User|test_results|user_executed_results|forbidden|none|0
User|test_runs|user_created_runs|forbidden|none|0
User|time_logs|user_time_logs|forbidden|none|0
User|user_blocks|blocked_id|forbidden|none|0
User|user_blocks|blocker_id|forbidden|none|0
`

const (
	inboundRequestFixtureExternalClosureManifestExpectedLineCount = 82
	inboundRequestFixtureExternalClosureManifestExpectedByteCount = 4767
	inboundRequestFixtureExternalClosureManifestExpectedSHA256    = "889c14127f47bdb73108bfef07573ff4b0bde7fd3b78b47122718a1f5b0a36b7"
)
