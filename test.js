try {
  $(new Date($file.info("mailboxes.db").lastModified))
} catch (err) {
  console.log(err)
}
