<template>
  <StaticPage
    title="The Sturdy command line application"
    metadescription="Learn how to use the Sturdy command line application (CLI)"
    category="help"
    image="https://images.unsplash.com/photo-1531403009284-440f080d1e12?ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&ixlib=rb-1.2.1&auto=format&fit=crop&w=1050&q=80"
  >
    <div class="text-base max-w-prose mx-auto lg:max-w-none">
      <p class="text-lg text-gray-500">
        The Sturdy command line application ("sturdy") manages a directory on your computer and
        connects it to Sturdy.
      </p>
    </div>
    <div
      class="mt-5 prose prose-yellow text-gray-500 mx-auto lg:max-w-none lg:row-start-1 lg:col-start-1"
    >
      <h2>Installation</h2>
      <p>
        Detailed installation instructions for macOS, Linux and Windows are available
        <a href="https://getsturdy.com/install">here</a> <i>(login required)</i>.
      </p>

      <h2>Setup a new directory</h2>
      <p>
        Run <code>sturdy init $TOKEN $PATH</code> to setup Sturdy in a new directory. There are
        detailed instructions available on the Codebase page on the Sturdy website. The
        <code>$PATH</code> must point to a empty or non-existant directory on your computer.
      </p>

      <h2>Starting and stopping Sturdy</h2>
      <p>
        The Sturdy application runs in the background, and can be started with
        <code>sturdy start</code> and stopped with <code>sturdy stop</code>.
      </p>
      <p>
        If Sturdy doesn't auto-start after restarting your computer, start Sturdy again with
        <code>sturdy start</code>
      </p>

      <h2>Logging in</h2>
      <p>
        Sturdy will prompt you to authenticate yourself the first time you run
        <code>sturdy init</code> automatically. To authenticate yourself as a different account, run
        <code>sturdy auth</code> and follow the instructions.
      </p>

      <h2>Moving a directory</h2>
      <p>
        It's currently not possible to move a connected Sturdy directory. Instead you can create a
        new directory, by following the "Setup Sturdy on your computer" instructions again for the
        Codebase to setup a new directory. The last argument to <code>sturdy init</code> is a path
        that you can change to the path where you want Sturdy to create the new directory. To delete
        the old directory, follow the delete instructions below.
      </p>
      <p>
        To move an existing Workspace to the new directory, use the "computer picker" in the top
        right of the page when viewing a Workspace, and select the new directory that you want to
        use. The workspace, and all unsaved changes will be moved automatically.
      </p>

      <h2>Deleting a Sturdy directory</h2>
      <p>
        To remove a directory from Sturdy, start by deleting the directory like you would delete any
        directory. Then open the <code>.sturdy</code> file in your <code>$HOME</code> and remove the
        section about this directory. Remember to make sure that the file you're saving is valid
        JSON, it's easy to miss a trailing comma!
      </p>

      <h2>Uninstalling Sturdy</h2>
      <p>
        Remove the <code>sturdy</code> and <code>sturdy-sync</code> binaries from your $PATH. Sturdy
        stores data in your home directory in the following files and directories:
      </p>
      <ul>
        <li>.sturdy</li>
        <li>.sturdy-agent</li>
        <li>.sturdy-sync</li>
      </ul>
      <p>Delete them to uninstall Sturdy.</p>
    </div>
  </StaticPage>
</template>

<script lang="ts" setup>
import StaticPage from '../../layouts/StaticPage.vue'
</script>
