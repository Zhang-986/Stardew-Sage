namespace go echofarmrpc

struct RelayRequest {
  1: required string request_id
  2: required string token
  3: optional string query
  4: optional binary body
  5: optional string content_type
}

struct RelayResponse {
  1: required string request_id
  2: required i32 status_code
  3: required string content_type
  4: required binary body
}

service EchoFarmGateway {
  RelayResponse Health(1: RelayRequest request)
  RelayResponse Learn(1: RelayRequest request)
  RelayResponse NextAction(1: RelayRequest request)
  RelayResponse ActionResult(1: RelayRequest request)
  RelayResponse Correction(1: RelayRequest request)
  RelayResponse PlayerModel(1: RelayRequest request)
  RelayResponse Skill(1: RelayRequest request)
  RelayResponse Memory(1: RelayRequest request)
  RelayResponse ModelUsage(1: RelayRequest request)
}
