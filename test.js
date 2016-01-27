try {
  $file.write('write', "hello");
  // $file.copy("write", "copy");
  // console.log("size is: " + $file.size("write"));
  // $file.write('moveme', 'move');
  // $file.move("moveme", "move");
  $file.mkdir("/Users/chris/Downloads/mkdir/test");
  // $file.write("deleteme", "delete");
  // $file.delete("deleteme");
  console.log("exists?: " + $file.exists("write"));
  console.log("read this: " + $file.readString("write"));
} catch(err) {
  throw(new Error(err));
} finally {
}
