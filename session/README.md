# session

Helper lib to allow easy setting and getting of a user's account on requests. The account is first attempted to be fetched from memcache. If an account is not found it is pulled from the datastore and cached.