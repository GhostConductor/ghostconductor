## How to Use Ghost Conductor

**1. Set your API key**
Go to **Providers** and set your API key. Keys are stored in memory, never in clear text — if the app restarts, you'll need to re-enter it.

**2. Add a repo**
Go to **Repos** and click **Add Repo**. Provide a name, URL, and branch. After adding, click into the repo card to set your GitHub token. Be sure to create a PAT scoped with minimal, secure permissions.

**3. Summon a ghost**
Go to **Summonings** and click **Conjure Spirit**:
   - Choose an intent
   - Describe the task
   - Pick your model/provider (only models with configured keys will appear)

**4. Review the PR**
When the ghost finishes, it opens a pull request. Review, request changes, or merge.

**5. Monitor usage**
Track token cost per job, per provider, and per model on the **Summonings** and **Providers** pages.
