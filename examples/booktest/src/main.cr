require "db"
require "pg"
require "./db"

# Example usage of the generated code
DB.open("postgres://localhost/booktest") do |db|
  queries = Db::Queries.new(db)
  
  # Create an author
  author = queries.create_author("Jane Doe", "A wonderful writer")
  if author
    puts "Created author: #{author.name} (ID: #{author.id})"
    
    # List all authors
    authors = queries.list_authors
    puts "\nAll authors:"
    authors.each do |a|
      puts "  #{a.id}: #{a.name}"
    end
    
    # Get a specific author
    if found = queries.get_author(author.id)
      puts "\nFound author: #{found.name}"
    end
  end
rescue ex
  puts "Error: #{ex.message}"
end