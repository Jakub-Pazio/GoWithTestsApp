import { useState, useEffect } from 'react';

interface Player {
  Name: string;
  Wins: number;
}

export default function ScoresView() {
  const [players, setPlayers] = useState<Player[]>([]);

  useEffect(() => {
    fetch('http://localhost:5050/league')
      .then(response => response.json())
      .then((data: Player[]) => setPlayers(data))
      .catch(error => console.error('Error fetching data:', error));
  }, []);

  return (
    <div>
      <h2>League Table</h2>
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Wins</th>
          </tr>
        </thead>
        <tbody>
          {players.map((player, index) => (
            <tr key={index}>
              <td>{player.Name}</td>
              <td>{player.Wins}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

