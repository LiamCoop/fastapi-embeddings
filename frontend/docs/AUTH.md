Users are authenticated with clerk. 

There is an orgSlug stored in the session token we get from clerk. However when a user first creates an account, we also need to create an organization for them.
Upon their first redirect to /dashboard, @papp/(app)/dashboard/route.ts checks to see if there's an org in the sessionClaims. 
 - If one is present, then we will redirect them to that org's page.
 - If none are present, a new org is created and we redirect user there.
