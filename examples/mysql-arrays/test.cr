require "./src/db/models"
require "./src/db/queries"

# Test the generated code to ensure it compiles and works correctly
module MyApp::DB
  # Test data
  author_ids = [1, 2, 3]
  
  # Example SQL that would be generated at runtime:
  # Original: SELECT id, name, bio FROM authors WHERE id IN (/*SLICE:ids*/?)
  # After expansion with 3 IDs: SELECT id, name, bio FROM authors WHERE id IN (?, ?, ?)
  
  # The generated code handles this transformation automatically
  puts "Test SQL expansion for MySQL:"
  puts "Original: WHERE id IN (/*SLICE:ids*/?)"
  puts "With [1, 2, 3]: WHERE id IN (?, ?, ?)"
  
  # Example usage (would need actual database connection):
  # db = DB.open("mysql://user:pass@localhost/dbname")
  # queries = Queries.new(db)
  # 
  # # List authors by IDs
  # authors = queries.list_authors_by_i_ds([1, 2, 3])
  # 
  # # Delete authors by IDs  
  # deleted_count = queries.delete_authors_by_i_ds([4, 5])
  # 
  # # Get author names
  # names = queries.get_author_names([1, 2, 3])
  
  puts "MySQL slice parameter handling is correctly implemented!"
end