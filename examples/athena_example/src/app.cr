require "athena"
require "db"
require "sqlite3"

require "./dto"
require "./db/database"

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
    schema = File.read("sqlc/schema.sql")
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
    Blog::UsersRepository.create(
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
    Blog::UsersRepository.find(id)
  end

  # Posts endpoints
  @[ARTA::Get("/posts")]
  def list_posts(limit : Int64 = 10, offset : Int64 = 0) : Array(Blog::GetPostRow)
    setup_database
    Blog::PostsRepository.all(limit: limit, offset: offset)
  end

  @[ARTA::Get("/posts/{slug}")]
  def get_post_by_slug(slug : String) : Blog::GetPostRow?
    setup_database
    Blog::PostsRepository.find_by_slug(slug)
  end

  @[ARTA::Post("/posts")]
  def create_post(request : ATH::Request) : Blog::Post?
    setup_database
    body_str = request.body.try &.gets_to_end
    raise ATH::Exception::BadRequest.new "Request body is empty." unless body_str

    body = PostCreateRequest.from_json(body_str)
    Blog::PostsRepository.create(
      user_id: body.user_id,
      title: body.title,
      slug: body.slug,
      content: body.content,
      excerpt: body.excerpt,
      published: body.published?,
      set_published_at: body.published? ? "1" : "0"
    )
  end

  @[ARTA::Post("/posts/{id}/publish")]
  def publish_post(id : Int64, user_id : Int64) : Nil
    setup_database
    Blog::PostsRepository.publish_post(id: id, user_id: user_id)
    nil
  end

  @[ARTA::Delete("/posts/{id}")]
  def delete_post(id : Int64, user_id : Int64) : Nil
    setup_database
    Blog::PostsRepository.delete(id: id, user_id: user_id)
    nil
  end

  # Tags endpoints
  @[ARTA::Get("/tags")]
  def list_tags : Array(Blog::Tag)
    setup_database
    Blog::TagsRepository.all
  end

  @[ARTA::Post("/tags")]
  def create_tag(request : ATH::Request) : Blog::Tag?
    setup_database
    body_str = request.body.try &.gets_to_end
    raise ATH::Exception::BadRequest.new "Request body is empty." unless body_str

    body = TagCreateRequest.from_json(body_str)
    Blog::TagsRepository.create(
      name: body.name,
      slug: body.slug,
      description: body.description
    )
  end

  @[ARTA::Get("/tags/{slug}/posts")]
  def get_posts_by_tag(slug : String, limit : Int64 = 10, offset : Int64 = 0) : Array(Blog::GetPostRow)
    setup_database
    tag = Blog::TagsRepository.find_by_slug(slug)
    return [] of Blog::GetPostRow unless tag

    Blog::PostsRepository.finds_by_tag(tag_id: tag.id, limit: limit, offset: offset)
  end
end

# Start the server
Athena::Framework.run
