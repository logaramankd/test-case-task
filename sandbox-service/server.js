import express from "express";
import { Sandbox } from "e2b";

const app = express();
app.use(express.json());

app.post("/run", async (req, res) => {
  try {
    const { code, input, language } = req.body;
    const lang = (language || "python").toLowerCase();

    const sandbox = await Sandbox.create("langsupport-dev");

    if (lang === "python") {
      // Write Python file
      await sandbox.commands.run(`cat << 'EOF' > solution.py
${code}
EOF`);

      // Run with input
      const result = await sandbox.commands.run(
        `echo "${input}" | python3 solution.py`
      );

      res.json({
        stdout: result.stdout,
        stderr: result.stderr,
        exitCode: result.exitCode,
      });
      return;
    }

    if (lang === "go") {
      // Write Go file
      await sandbox.commands.run(`cat << 'EOF' > main.go
${code}
EOF`);

      // Run with input
      const result = await sandbox.commands.run(
        `echo "${input}" | go run main.go`
      );

      res.json({
        stdout: result.stdout,
        stderr: result.stderr,
        exitCode: result.exitCode,
      });
      return;
    }

    if (lang === "javascript" || lang === "js" || lang === "node") {
      // Write JavaScript file
      await sandbox.commands.run(`cat << 'EOF' > solution.js
${code}
EOF`);

      // Run with input piped to the script (process.stdin)
      const result = await sandbox.commands.run(
        `echo "${input}" | node solution.js`
      );

      res.json({
        stdout: result.stdout,
        stderr: result.stderr,
        exitCode: result.exitCode,
      });
      return;
    }
    if (lang === "java") {
      // Write Java file
      await sandbox.commands.run(`cat << 'EOF' > Main.java
${code}
EOF`);

      // Compile
      const compile = await sandbox.commands.run(`javac Main.java`);

      if (compile.exitCode !== 0) {
        res.json({
          stdout: "",
          stderr: compile.stderr,
          exitCode: compile.exitCode,
        });
        return;
      }

      // Run with input
      const result = await sandbox.commands.run(
        `echo "${input}" | java Main`
      );

      res.json({
        stdout: result.stdout,
        stderr: result.stderr,
        exitCode: result.exitCode,
      });
      return;
    }

    res
      .status(400)
      .json({ error: "Unsupported language. Use python, go, or javascript." });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.listen(3001, () => {
  console.log("Sandbox service running on port 3001");
});
