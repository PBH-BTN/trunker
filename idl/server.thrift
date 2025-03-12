include "./common.thrift"
namespace go pbh.btn.trunker

service TrunkerService{
    AnnounceResponse Announce(1:AnnounceRequest request)
    ScrapeResponse Scrape(1:ScrapeRequest request)
    GetStatisticResponse GetStatistic(1:GetStatisticRequest request)
    BanResponse Ban(1: BanRequest request)
    DeleteInfoHashResponse DeleteInfoHash(1: DeleteInfoHashRequest request)
    GetPeerResponse GetPeer(1: GetPeerRequest request)
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

struct ScrapeRequest{
    1: required list<string> info_hashes
}

struct ScrapeResponse{
    1: required map<string,common.ScrapeFile> res
}

struct GetStatisticRequest{
}

struct GetStatisticResponse{
    1: required common.StatisticInfo info
}

enum BanType{
    InfoHash = 0
    PeerID = 1
}

struct BanRequest {
    1: required BanType type
    2: required string target
}

struct BanResponse{}


struct DeleteInfoHashRequest{
    1: required string target
}

struct DeleteInfoHashResponse{}

struct GetPeerRequest{
    1: required string info_hash
}

struct GetPeerResponse{
    1: required list<common.Peer> peers
}