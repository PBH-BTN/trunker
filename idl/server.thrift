include "./common.thrift"
namespace go pbh.btn.trunker

service TrunkerService{
    AnnounceResponse Announce(1:AnnounceRequest request)
}

struct AnnounceRequest{
    1: optional list<common.Offer> offers
    2: required binary client_ip
    3: required string info_hash;
    4: required string peer_id;
    5: required string ip
    6: string ipv4
    7: string ipv6
    8: required i64 downloaded
    9: required i64 left
    10: required i64 uploaded
    11: required i64 num_want
    12: required common.PeerType type
    13: required common.Source source
    14: required common.PeerEvent event
    15: required i32 port
    16: bool compact

}

struct AnnounceResponse{
    1: required list<common.Peer> peers
}