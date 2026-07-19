package constants

// Contains constants related to UI formatting conventions

const BotName string = "JacuzziBot"
const UIFloatMaxDisplayedDecimalPrecision int = 2

// Error messages that show up in multiple places:
const ErrorReportMessageSuffix string = ". Send a screenshot to one of your admins and they'll fix this."
const ErrorFetchingDisplayNameMessage string = "(unable to fetch display name, whoops!)"

// Sets of runes (regex-set-friendly) that can be used to sanitize inputs. We want to allow common sentence
// formatting runes like ,.-; so they're not listed here.
const MarkdownRegex = "*_~#\\[\\]`|>"
const NewLineRegex = "\\n\\r"
const DiscordApiRegex = "@<>:"
const SlashRegex = "/\\\\"
const OtherSpecialsRegex = "\"{}()"

const InputDenylist = MarkdownRegex + NewLineRegex + DiscordApiRegex + SlashRegex + OtherSpecialsRegex
const InputDenylistRegex = ".*[" + InputDenylist + "].*"
const InputDenylistForPrinting = "*_~#[]{}()`|>@<>:/\\\""
