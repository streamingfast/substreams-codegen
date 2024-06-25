import { log } from "@graphprotocol/graph-ts";
import { Protobuf } from "as-proto/assembly";
import { Attr, Event } from "../generated/schema";
import { EventList } from "./pb/sf/substreams/cosmos/v1/EventList";

export function handleEvents(bytes: Uint8Array): void {
  const eventList: EventList = Protobuf.decode<EventList>(
    bytes,
    EventList.decode
  );
  const events = eventList.events;

  log.info("Protobuf decoded, length: {}", [events.length.toString()]);

  for (let i = 0; i < events.length; i++) {
    const event = events[i].event;
    if (event == null) {
      continue;
    }

    const trxHash = events[i].transactionHash;
    const eventID = trxHash + "-" + i.toString();

    let attributes: string[] = [];
    for (let i = 0; i < event.attributes.length; ++i) {
      const attribute = event.attributes[i];

      const key = attribute.key;
      const value = attribute.value;

      const attributeEntity = new Attr(eventID + "-" + i.toString());
      attributeEntity.key = key;
      attributeEntity.value = value;
      attributeEntity.save();
      attributes.push(attributeEntity.id);
    }

    let entity = new Event(eventID); // need to set an id
    entity.type = event.type;
    entity.attrs = attributes;

    entity.save();
    log.debug("Entity saved: {}", [entity.id]);
  }
}
