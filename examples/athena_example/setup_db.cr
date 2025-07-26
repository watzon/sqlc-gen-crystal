#!/usr/bin/env crystal

require "db"
require "sqlite3"

# Simple database setup script for the blog example
def setup_database
  puts "Setting up blog database..."
  
  # Remove existing database if it exists
  File.delete("blog.db") if File.exists?("blog.db")
  
  # Create new database and run schema
  DB.open("sqlite3:./blog.db") do |db|
    schema = File.read("sqlc/schema.sql")
    
    # Split schema into individual statements and execute them
    statements = schema.split(';').map(&.strip).reject(&.empty?)
    
    puts "Found #{statements.size} statements to execute"
    statements.each_with_index do |statement, i|
      # Remove comment lines from each statement
      cleaned = statement.lines.reject { |line| line.strip.starts_with?("--") }.join("\n").strip
      next if cleaned.empty?
      
      puts "Executing statement #{i + 1}: #{cleaned[0..80]}..."
      db.exec(cleaned)
    end
    puts "✓ Database schema created"
    
    # Add some sample data
    puts "Adding sample data..."
    
    # Create a sample user
    user_id = db.query_one("INSERT INTO users (username, email, password_hash, display_name, bio) VALUES (?, ?, ?, ?, ?) RETURNING id",
      "john_doe", 
      "john@example.com", 
      "hashed_password_123", 
      "John Doe", 
      "A Crystal enthusiast and blogger"
    ) { |rs| rs.read(Int64) }
    puts "✓ Created user: john_doe (id: #{user_id})"
    
    # Create sample tags
    crystal_tag_id = db.query_one("INSERT INTO tags (name, slug, description) VALUES (?, ?, ?) RETURNING id",
      "Crystal", 
      "crystal", 
      "Posts about the Crystal programming language"
    ) { |rs| rs.read(Int64) }
    
    web_tag_id = db.query_one("INSERT INTO tags (name, slug, description) VALUES (?, ?, ?) RETURNING id",
      "Web Development", 
      "web-development", 
      "Posts about web development and frameworks"
    ) { |rs| rs.read(Int64) }
    puts "✓ Created tags: Crystal, Web Development"
    
    # Create sample posts
    post1_id = db.query_one("INSERT INTO posts (user_id, title, slug, content, excerpt, published, published_at) VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id",
      user_id,
      "Getting Started with Crystal",
      "getting-started-with-crystal",
      "Crystal is a programming language that combines the efficiency of C with the expressiveness of Ruby. In this post, we'll explore the basics of Crystal programming...",
      "Learn the fundamentals of Crystal programming language",
      true,
      Time.utc
    ) { |rs| rs.read(Int64) }
    
    post2_id = db.query_one("INSERT INTO posts (user_id, title, slug, content, excerpt, published, published_at) VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id",
      user_id,
      "Building Web APIs with Athena Framework",
      "building-web-apis-with-athena",
      "The Athena Framework provides a powerful foundation for building web APIs in Crystal. This tutorial covers routing, controllers, and database integration...",
      "Learn how to build robust web APIs using Athena Framework",
      true,
      Time.utc
    ) { |rs| rs.read(Int64) }
    
    post3_id = db.query_one("INSERT INTO posts (user_id, title, slug, content, excerpt, published) VALUES (?, ?, ?, ?, ?, ?) RETURNING id",
      user_id,
      "Type-Safe Database Queries with SQLC",
      "type-safe-database-queries-sqlc",
      "SQLC generates type-safe database code from SQL queries. Combined with Crystal's type system, this provides compile-time safety for database operations...",
      "Explore type-safe database programming with SQLC and Crystal",
      false
    ) { |rs| rs.read(Int64) }
    puts "✓ Created 3 sample posts (2 published, 1 draft)"
    
    # Associate posts with tags
    db.exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)", post1_id, crystal_tag_id)
    db.exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)", post2_id, crystal_tag_id)
    db.exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)", post2_id, web_tag_id)
    db.exec("INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)", post3_id, crystal_tag_id)
    puts "✓ Associated posts with tags"
    
    puts "\n🎉 Database setup complete!"
    puts "\nSample data created:"
    puts "  • User: john_doe (john@example.com)"
    puts "  • Tags: Crystal, Web Development"
    puts "  • Posts: 3 posts (2 published, 1 draft)"
    puts "\nYou can now start the server with: ./bin/app"
  end
rescue ex
  puts "❌ Error setting up database: #{ex.message}"
  exit(1)
end

setup_database