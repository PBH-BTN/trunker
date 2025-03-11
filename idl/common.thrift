namespace go pbh.btn.trunker
enum PeerEvent{
    unknown = 0,
    started = 1,
    stopped = 2,
    completed = 3,
}

enum PeerType {
    Bittorrent = 0,
    Webtorrent = 1,
}

enum Source{
    HTTP = 0
    UDP = 1
    WS = 2
}

struct Offer{
   1: required string offer_id
   2: required OfferDetail offer
}

struct OfferDetail{
    1: required string type;
    2: required string sdp
}

struct Peer{
    1: required binary ip
    2: binary ipv4
    3: binary ipv6
    4: binary client_ip
    5: required i64 last_seen
    6: optional list<Offer> offers
    7: required string id
    8: required string user_agent
    9: required i32 port
    10: required i64 uploaded
    11: required i64 downloaded
    12: required i64 left
    13: required PeerType type
    14: required PeerEvent event
    15: required Source source
}