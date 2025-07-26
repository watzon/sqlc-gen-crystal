require "athena"
require "db"
require "sqlite3"
require "./db/*"
require "./db/repositories/*"

# Simple blog API built with Athena Framework and sqlc-gen-crystal
class BlogController < ATH::Controller
  # Initialize database connection using the connection manager
  private def setup_database
    ENV["DATABASE_URL"] = "sqlite3:./blog.db"
  end

  # Initialize database
  @[ARTA::Get("/init")]
  def init_db : String
    setup_database
    schema = File.read("schema.sql")
    Blog::Database.connection.exec(schema)
    "Database initialized!"
  end

  # Users endpoints
  @[ARTA::Post("/users")]
  def create_user(request : ATH::Request) : Blog::User?
    setup_database
    body_str = request.body.try &.gets_to_end
    raise ATH::Exception::BadRequest.new "Request body is empty." unless body_str

    body = UserCreateRequest.from_json(body_str)
    users_repo = Blog::UsersRepository.new
    users_repo.create(
      username: body.username,
      email: body.email,
      password_hash: body.password_hash,
      display_name: body.display_name,
      bio: body.bio
    )
  end

  @[ARTA::Get("/users/{id}")]
  def get_user(id : Int64) : Blog::User?
    setup_database
    users_repo = Blog::UsersRepository.new
    users_repo.find(id)
  end

  # Posts endpoints
  @[ARTA::Get("/posts")]
  def list_posts(limit : Int64 = 10, offset : Int64 = 0) : Array(Blog::GetPostRow)
    setup_database
    posts_repo = Blog::PostsRepository.new
    posts_repo.all(limit, offset)
  end

  @[ARTA::Get("/posts/{slug}")]
  def get_post_by_slug(slug : String) : Blog::GetPostRow?
    setup_database
    posts_repo = Blog::PostsRepository.new
    posts_repo.find_by_slug(slug)
  end

  @[ARTA::Post("/posts")]
  def create_post(request : ATH::Request) : Blog::Post?
    setup_database
    body_str = request.body.try &.gets_to_end
    raise ATH::Exception::BadRequest.new "Request body is empty." unless body_str

    body = PostCreateRequest.from_json(body_str)
    posts_repo = Blog::PostsRepository.new
    posts_repo.create(
      body.user_id,
      body.title,
      body.slug,
      body.content,
      body.excerpt,
      body.published?,
      body.published? ? "1" : "0" # This is the arg7 parameter for the CASE WHEN clause
    )
  end

  @[ARTA::Post("/posts/{id}/publish")]
  def publish_post(id : Int64, user_id : Int64) : Nil
    setup_database
    posts_repo = Blog::PostsRepository.new
    posts_repo.publish_post(id, user_id)
    nil
  end

  @[ARTA::Delete("/posts/{id}")]
  def delete_post(id : Int64, user_id : Int64) : Nil
    setup_database
    posts_repo = Blog::PostsRepository.new
    posts_repo.delete(id, user_id)
    nil
  end

  # Tags endpoints
  @[ARTA::Get("/tags")]
  def list_tags : Array(Blog::Tag)
    setup_database
    tags_repo = Blog::TagsRepository.new
    tags_repo.all
  end

  @[ARTA::Post("/tags")]
  def create_tag(request : ATH::Request) : Blog::Tag?
    setup_database
    body_str = request.body.try &.gets_to_end
    raise ATH::Exception::BadRequest.new "Request body is empty." unless body_str

    body = TagCreateRequest.from_json(body_str)
    tags_repo = Blog::TagsRepository.new
    tags_repo.create(
      name: body.name,
      slug: body.slug,
      description: body.description
    )
  end

  @[ARTA::Get("/tags/{slug}/posts")]
  def get_posts_by_tag(slug : String, limit : Int64 = 10, offset : Int64 = 0) : Array(Blog::GetPostRow)
    setup_database
    tags_repo = Blog::TagsRepository.new
    posts_repo = Blog::PostsRepository.new
    
    tag = tags_repo.find_by_slug(slug)
    return [] of Blog::GetPostRow unless tag

    posts_repo.finds_by_tag(tag.id, limit, offset)
  end
end

# Request DTOs
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

# Start the server
Athena::Framework.run
