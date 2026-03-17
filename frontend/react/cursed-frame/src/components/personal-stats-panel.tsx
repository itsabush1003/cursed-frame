const PersonalStatsPanel = ({
  teamOrder,
  personalOrder,
  correctRate,
}: {
  teamOrder: number;
  personalOrder: number;
  correctRate: number;
}) => {
  return (
    <div>
      <h3>あなたのチームは{teamOrder}位でした</h3>
      <p>個人順位： {personalOrder}</p>
      <p>正解率： {(correctRate * 100).toFixed(1) + "%"}</p>
    </div>
  );
};

export default PersonalStatsPanel;
