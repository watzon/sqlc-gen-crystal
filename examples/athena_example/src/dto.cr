struct UserCreateRequest
  include JSON::Serializable

  getter username : String
  getter email : String
  getter password_hash : String
  getter display_name : String?
  getter bio : String?
end

struct PostCreateRequest
  include JSON::Serializable

  getter user_id : Int64
  getter title : String
  getter slug : String
  getter content : String
  getter excerpt : String?
  getter? published : Bool = false
end

struct TagCreateRequest
  include JSON::Serializable

  getter name : String
  getter slug : String
  getter description : String?
end
