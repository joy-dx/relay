package dto

type RlyOpts struct {
	Channel EventChannel
	Type    EventRef
}

type relayEventWithOpts struct {
	RelayEventInterface

	channel EventChannel
	typ     EventRef
}

func (e relayEventWithOpts) RelayChannel() EventChannel {
	if e.channel != "" {
		return e.channel
	}

	return e.RelayEventInterface.RelayChannel()
}

func (e relayEventWithOpts) RelayType() EventRef {
	if e.typ != "" {
		return e.typ
	}

	return e.RelayEventInterface.RelayType()
}

func ApplyRelayOpts(
	event RelayEventInterface,
	opts RlyOpts,
) RelayEventInterface {
	if opts.Channel == "" && opts.Type == "" {
		return event
	}

	return relayEventWithOpts{
		RelayEventInterface: event,
		channel:             opts.Channel,
		typ:                 opts.Type,
	}
}
