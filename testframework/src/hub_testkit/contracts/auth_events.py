from __future__ import annotations

from typing import NoReturn

from .source_bundle import ContractSourceBundle, ContractSourcePath


_MAX_FUNCTION_CHARS = 256 * 1024
_MAX_FUNCTION_LINES = 4_096
_MAX_SOURCE_CHARS = 4 * 1024 * 1024
_INVALID = object()


class AuthEventContractError(ValueError):
    pass


def assert_auth_session_created_event_contract(bundle: object) -> None:
    result = _auth_session_created_event_valid_safe(bundle)
    del bundle
    if result is True:
        return
    del result
    _raise_invalid()


def _auth_session_created_event_valid_safe(bundle: object) -> bool:
    try:
        if type(bundle) is not ContractSourceBundle:
            return False
        event_source = bundle.select(ContractSourcePath.AUTH_EVENTS)
        producer_source = bundle.select(ContractSourcePath.AUTH_SERVICE_SESSION)
        mapping_source = bundle.select(ContractSourcePath.NOTIFICATION_EVENTS)
        consumer_source = bundle.select(ContractSourcePath.NOTIFICATION_CONSUMER)
        event_text = event_source.decode("utf-8", errors="strict")
        producer_text = producer_source.decode("utf-8", errors="strict")
        mapping_text = mapping_source.decode("utf-8", errors="strict")
        consumer_text = consumer_source.decode("utf-8", errors="strict")
        return (
            _event_shape_valid(event_text)
            and _producer_link_valid(producer_text)
            and _notification_mapping_valid(mapping_text)
            and _typed_consumer_link_valid(consumer_text)
        )
    except Exception:
        return False


def _event_shape_valid(source: str) -> bool:
    lines = _top_level_block_lines(
        source,
        "type SessionCreatedEvent struct {",
        opening="{",
        closing="}",
    )
    event_shape = (
        "type SessionCreatedEvent struct {",
        'UserID    string `json:"user_id"`',
        'SessionID string `json:"session_id"`',
        'Device    string `json:"device"`',
        'IPAddress string `json:"ip_address"`',
        "}",
    )
    return type(lines) is tuple and lines == event_shape


def _producer_link_valid(source: str) -> bool:
    create_lines = _function_lines(
        source,
        "func (s *SessionService) CreateSession(ctx context.Context, "
        "userID uuid.UUID, device, ip string) (*TokenPairResponse, error) {",
    )
    publish_lines = _function_lines(
        source,
        "func (s *SessionService) publishSessionCreated("
        "userID, sessionID, device, ip string) {",
    )
    if type(create_lines) is not tuple or type(publish_lines) is not tuple:
        return False
    publish_call = (
        "s.publishSessionCreated(userID.String(), sessionID.String(), device, ip)",
        "return &TokenPairResponse{AccessToken: accessToken, "
        "RefreshToken: refreshToken, "
        "ExpiresIn: int64(s.accessTTL.Seconds())}, nil",
    )
    if _contiguous_count(create_lines, publish_call) != 1:
        return False
    if create_lines.count(publish_call[0]) != 1:
        return False
    publish_expected = (
        "data, err := json.Marshal(SessionCreatedEvent{UserID: userID, "
        "SessionID: sessionID, Device: device, IPAddress: ip})",
        'if err := s.nats.Publish("auth.session.created", data); err != nil {',
    )
    return _ordered_unique_body_lines(publish_lines, publish_expected)


def _notification_mapping_valid(source: str) -> bool:
    const_blocks = _top_level_block_groups(
        source,
        "const (",
        opening="(",
        closing=")",
    )
    function_lines = _function_lines(source, "func SubjectToType(subject string) string {")
    if not isinstance(const_blocks, tuple) or not isinstance(function_lines, tuple):
        return False
    declaration = 'TypeNewLogin = "new_login"'
    if sum(block.count(declaration) for block in const_blocks) != 1:
        return False
    return _contiguous_count(
        function_lines,
        (
            'if subject == "auth.session.created" {',
            "return TypeNewLogin",
            "}",
        ),
    ) == 1


def _typed_consumer_link_valid(source: str) -> bool:
    process_lines = _function_lines(
        source,
        "func (c *Consumer) processEvent(ctx context.Context, msg jetstream.Msg) {",
    )
    handler_lines = _function_lines(
        source,
        "func (c *Consumer) handleNewLogin(ctx context.Context, data []byte) error {",
    )
    if type(process_lines) is not tuple or type(handler_lines) is not tuple:
        return False
    process_expected = (
        "subject := msg.Subject()",
        "notifType := SubjectToType(subject)",
        "case TypeNewLogin:",
        "processErr = c.handleNewLogin(ctx, msg.Data())",
    )
    if not _ordered_unique_body_lines(process_lines, process_expected):
        return False
    if _contiguous_count(
        process_lines,
        (
            "case TypeNewLogin:",
            "processErr = c.handleNewLogin(ctx, msg.Data())",
        ),
    ) != 1:
        return False
    typed_unmarshal = (
        "var evt auth.SessionCreatedEvent",
        "if err := json.Unmarshal(data, &evt); err != nil {",
    )
    if _contiguous_count(handler_lines, typed_unmarshal) != 1:
        return False
    dispatch_start = "if err := c.service.Dispatch(ctx, DispatchEvent{"
    if not _ordered_unique_body_lines(
        handler_lines,
        (typed_unmarshal[1], dispatch_start),
    ):
        return False
    return _dispatch_block_valid(handler_lines, dispatch_start)


def _dispatch_block_valid(lines: tuple[str, ...], start: str) -> bool:
    end = "}); err != nil {"
    if lines.count(start) != 1 or lines.count(end) != 1:
        return False
    start_index = lines.index(start)
    end_index = lines.index(end)
    if end_index <= start_index or end_index - start_index > 32:
        return False
    block = lines[start_index : end_index + 1]
    event_type = "Type:       TypeNewLogin,"
    entity_id = "EntityID:   evt.SessionID,"
    if block.count(event_type) != 1 or block.count(entity_id) != 1:
        return False
    if sum(line.startswith("Type:") for line in block) != 1:
        return False
    if sum(line.startswith("EntityID:") for line in block) != 1:
        return False
    return block.index(event_type) < block.index(entity_id)


def _ordered_unique_body_lines(
    lines: tuple[str, ...],
    expected: tuple[str, ...],
) -> bool:
    if any(lines.count(line) != 1 for line in expected):
        return False
    positions = tuple(lines.index(line) for line in expected)
    return positions == tuple(sorted(positions))


def _function_lines(source: str, signature: str) -> tuple[str, ...] | object:
    return _top_level_block_lines(source, signature, opening="{", closing="}")


def _extract_top_level_function(source: str, signature: str) -> str | object:
    blocks = _scan_top_level_blocks(source, signature, opening="{", closing="}")
    if not isinstance(blocks, tuple) or len(blocks) != 1:
        return _INVALID
    return blocks[0][0]


def _top_level_block_lines(
    source: str,
    signature: str,
    *,
    opening: str,
    closing: str,
) -> tuple[str, ...] | object:
    blocks = _scan_top_level_blocks(
        source,
        signature,
        opening=opening,
        closing=closing,
    )
    if not isinstance(blocks, tuple) or len(blocks) != 1:
        return _INVALID
    return blocks[0][1]


def _top_level_block_groups(
    source: str,
    signature: str,
    *,
    opening: str,
    closing: str,
) -> tuple[tuple[str, ...], ...] | object:
    blocks = _scan_top_level_blocks(
        source,
        signature,
        opening=opening,
        closing=closing,
    )
    if not isinstance(blocks, tuple) or not blocks:
        return _INVALID
    return tuple(block[1] for block in blocks)


def _scan_top_level_blocks(
    source: str,
    signature: str,
    *,
    opening: str,
    closing: str,
) -> tuple[tuple[str, tuple[str, ...]], ...] | object:
    if type(source) is not str or type(signature) is not str:
        return _INVALID
    if (
        len(source) > _MAX_SOURCE_CHARS
        or type(opening) is not str
        or type(closing) is not str
        or len(opening) != 1
        or len(closing) != 1
        or opening == closing
        or not signature.endswith(opening)
    ):
        return _INVALID
    brace_depth = 0
    state = "normal"
    escaped = False
    line_start = True
    block_start: int | None = None
    block_depth = 0
    code_lines: list[list[str]] | None = None
    literal_chars: list[str] | None = None
    literal_line = -1
    literal_multiline = False
    blocks: list[tuple[str, tuple[str, ...]]] = []
    index = 0
    while index < len(source):
        char = source[index]
        following = source[index + 1] if index + 1 < len(source) else ""
        if state == "normal" and brace_depth == 0 and line_start:
            signature_end = index + len(signature)
            if source.startswith(signature, index) and (
                signature_end == len(source) or source[signature_end] == "\n"
            ):
                if block_start is not None:
                    return _INVALID
                block_start = index
                block_depth = 0
                code_lines = [[]]
        capturing = block_start is not None
        if state == "line-comment":
            if char == "\n":
                if capturing and code_lines is not None:
                    code_lines.append([])
                state = "normal"
        elif state == "block-comment":
            if char == "\n" and capturing and code_lines is not None:
                code_lines.append([])
            if char == "*" and following == "/":
                state = "normal"
                index += 1
        elif state in {"double", "single"}:
            if char == "\n":
                return _INVALID
            if capturing and literal_chars is not None:
                literal_chars.append(char)
            if escaped:
                escaped = False
            elif char == "\\":
                escaped = True
            elif (state == "double" and char == '"') or (
                state == "single" and char == "'"
            ):
                state = "normal"
                if (
                    capturing
                    and code_lines is not None
                    and literal_chars is not None
                ):
                    code_lines[literal_line].append("".join(literal_chars))
                    literal_chars = None
                    literal_line = -1
        elif state == "raw":
            if capturing and literal_chars is not None:
                literal_chars.append(char)
            if char == "\n" and capturing and code_lines is not None:
                literal_multiline = True
                code_lines.append([])
            if char == "`":
                state = "normal"
                if (
                    capturing
                    and code_lines is not None
                    and literal_chars is not None
                ):
                    literal = "`<raw>`" if literal_multiline else "".join(literal_chars)
                    code_lines[literal_line].append(literal)
                    literal_chars = None
                    literal_line = -1
                    literal_multiline = False
        elif char == "/" and following == "/":
            if capturing and code_lines is not None:
                code_lines[-1].append(" ")
            state = "line-comment"
            index += 1
        elif char == "/" and following == "*":
            if capturing and code_lines is not None:
                code_lines[-1].append(" ")
            state = "block-comment"
            index += 1
        elif char == '"':
            state = "double"
            escaped = False
            if capturing and code_lines is not None:
                literal_chars = [char]
                literal_line = len(code_lines) - 1
                literal_multiline = False
        elif char == "'":
            state = "single"
            escaped = False
            if capturing and code_lines is not None:
                literal_chars = [char]
                literal_line = len(code_lines) - 1
                literal_multiline = False
        elif char == "`":
            state = "raw"
            if capturing and code_lines is not None:
                literal_chars = [char]
                literal_line = len(code_lines) - 1
                literal_multiline = False
        else:
            if capturing and code_lines is not None:
                if char == "\n":
                    code_lines.append([])
                else:
                    code_lines[-1].append(char)
            if char == "{":
                brace_depth += 1
            elif char == "}":
                brace_depth -= 1
                if brace_depth < 0:
                    return _INVALID
            if capturing and char == opening:
                block_depth += 1
            elif capturing and char == closing:
                block_depth -= 1
                if block_depth < 0:
                    return _INVALID
                if block_depth == 0:
                    if block_start is None or code_lines is None:
                        return _INVALID
                    block = source[block_start : index + 1]
                    lines = tuple(
                        line
                        for parts in code_lines
                        if (line := "".join(parts).strip())
                    )
                    if (
                        len(block) > _MAX_FUNCTION_CHARS
                        or len(code_lines) > _MAX_FUNCTION_LINES
                    ):
                        return _INVALID
                    blocks.append((block, lines))
                    block_start = None
                    code_lines = None
                    literal_chars = None
                    literal_line = -1
                    literal_multiline = False
        if char == "\n":
            line_start = True
        elif line_start:
            line_start = False
        index += 1
    if (
        brace_depth != 0
        or state not in {"normal", "line-comment"}
        or block_start is not None
        or block_depth != 0
        or literal_chars is not None
    ):
        return _INVALID
    return tuple(blocks)


def _contiguous_count(lines: tuple[str, ...], expected: tuple[str, ...]) -> int:
    width = len(expected)
    return sum(
        lines[index : index + width] == expected
        for index in range(len(lines) - width + 1)
    )


def _raise_invalid() -> NoReturn:
    raise AuthEventContractError("auth event source contract is invalid") from None
